package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aitumik/snippetbox/pkg/forms"
	"github.com/aitumik/snippetbox/pkg/models"
	"github.com/go-chi/chi/v5"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	s, err := app.snippet.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	expiring, err := app.snippet.GetExpiringToday()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := &TemplateData{Snippets: s, ExpiringToday: expiring}
	app.render(w, r, "home.page.tmpl", data)
}

func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")

	if !form.Valid() {
		tags, _ := app.tags.GetAll()
		data := &TemplateData{Form: form, Tags: tags}
		app.render(w, r, "create.page.tmpl", data)
		return
	}

	var tagIDs []int

	selectedTags := r.PostForm["tags"]
	for _, t := range selectedTags {
		t = strings.TrimSpace(t)
		if t != "" {
			tag, err := app.tags.GetByName(t)
			if err == models.ErrNoRecord {
				id, err := app.tags.Insert(t)
				if err != nil {
					app.serverError(w, err)
					return
				}
				tagIDs = append(tagIDs, id)
			} else if err != nil {
				app.serverError(w, err)
				return
			} else {
				tagIDs = append(tagIDs, tag.ID)
			}
		}
	}

	newTags := r.Form.Get("new_tags")
	if newTags != "" {
		for _, t := range strings.Split(newTags, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tag, err := app.tags.GetByName(t)
				if err == models.ErrNoRecord {
					id, err := app.tags.Insert(t)
					if err != nil {
						app.serverError(w, err)
						return
					}
					tagIDs = append(tagIDs, id)
				} else if err != nil {
					app.serverError(w, err)
					return
				} else {
					tagIDs = append(tagIDs, tag.ID)
				}
			}
		}
	}

	id, err := app.snippet.Insert(form.Get("title"), form.Get("content"), form.Get("expires"), tagIDs, app.authenticatedUser(r).ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Snippet created successfully!")
	http.Redirect(w, r, fmt.Sprintf("/snippets/%d", id), http.StatusSeeOther)
}

func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	tags, err := app.tags.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := &TemplateData{Form: forms.New(nil), Tags: tags}
	app.render(w, r, "create.page.tmpl", data)
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	s, err := app.snippet.Get(id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	data := &TemplateData{Snippet: s}
	app.render(w, r, "show.page.tmpl", data)
}

func (app *application) showTagSnippets(w http.ResponseWriter, r *http.Request) {
	tagName := chi.URLParam(r, "name")

	tag, err := app.tags.GetByName(tagName)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	snippets, err := app.snippet.GetByTag(tag.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := &TemplateData{
		Snippets: snippets,
		Tag:      tag,
	}
	app.render(w, r, "tag.page.tmpl", data)
}

func (app *application) showUserProfile(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	user, err := app.users.Get(id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	snippets, err := app.snippet.GetByUser(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := &TemplateData{
		User:     user,
		Snippets: snippets,
	}
	app.render(w, r, "user.page.tmpl", data)
}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &TemplateData{
		Form: forms.New(nil),
	})
}

func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)

	if !form.Valid() {
		data := &TemplateData{Form: form}
		app.render(w, r, "signup.page.tmpl", data)
		return
	}

	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err != nil {
		form.Errors.Add("email", "Email address already in use")
		data := &TemplateData{Form: form}
		app.render(w, r, "signup.page.tmpl", data)
		return
	}

	app.session.Put(r, "flash", "Registration successful. Please login.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &TemplateData{
		Form: forms.New(nil),
	})
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))

	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Wrong email or password")
		app.render(w, r, "login.page.tmpl", &TemplateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "userID", id)
	http.Redirect(w, r, "/snippets/create", http.StatusSeeOther)
}

func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	app.session.Remove(r, "userID")
	app.session.Put(r, "flash", "You have been logged out successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) extendExpiry(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	err = app.snippet.ExtendExpiry(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Snippet expiry extended by 1 day!")
	user := app.authenticatedUser(r)
	http.Redirect(w, r, fmt.Sprintf("/user/%d", user.ID), http.StatusSeeOther)
}

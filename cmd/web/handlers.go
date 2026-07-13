package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aitumik/snippetbox/pkg/forms"
	"github.com/aitumik/snippetbox/pkg/models"
	"github.com/go-chi/chi/v5"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// pat now handles request to the `/` route
	s, err := app.snippet.Latest()
	if err != nil {
		app.serverError(w, err)
	}
	// create a value of struct TemplateData to hold  slice of snippets
	data := &TemplateData{Snippets: s}
	app.render(w, r, "home.page.tmpl", data)
}

func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	// First wee need to call the method r.ParseForm() which loads the values of the post
	// to the r.PostForm map
	// We can get for example title if we do this `r.ParseForm().Get("Title")`
	// Note that the r.ParseForm() is limited to 10MB
	// To change this limit use the http.MaxBytesReader()
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

	// if the form is not valid redisplay the form
	if !form.Valid() {
		data := &TemplateData{
			Form: form,
		}
		app.render(w, r, "create.page.tmpl", data)
		return
	}

	id, err := app.snippet.Insert(form.Get("title"), form.Get("content"), form.Get("expires"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r, "flash", "Snippet created successfully!")
	http.Redirect(w, r, fmt.Sprintf("/snippets/%d", id), http.StatusSeeOther)
}

func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	data := &TemplateData{
		Form: forms.New(nil),
	}
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
	// Create and instance of a TemplateData struct holding the snippet data
	data := &TemplateData{
		Snippet: s,
	}

	app.render(w, r, "show.page.tmpl", data)
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
		data := &TemplateData{
			Form: form,
		}
		app.render(w, r, "signup.page.tmpl", data)
		return
	}

	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err != nil {
		form.Errors.Add("email", "Email address already in use")

		data := &TemplateData{
			Form: form,
		}
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

	// check if the credentials are valid if not add
	// generic error message to the form
	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))

	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "Wrong email or password")
		app.render(w, r, "login.page.tmpl", &TemplateData{
			Form: form,
		})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// Add the current id of the user to the session
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

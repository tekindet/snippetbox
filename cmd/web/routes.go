package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(app.recoverPanic)
	r.Use(app.logRequest)
	r.Use(secureHeaders)

	fileServer := http.FileServer(http.Dir(app.cfg.StaticDir))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	r.Get("/ping", ping)

	r.Group(func(r chi.Router) {
		r.Use(app.session.Enable)
		r.Use(noSurf)
		r.Use(app.authenticate)

		r.Get("/", app.home)
		r.Get("/snippets/{id}", app.showSnippet)
		r.Get("/tags/{name}", app.showTagSnippets)
		r.Get("/user/{id}", app.showUserProfile)

		r.Get("/user/signup", app.signupUserForm)
		r.Post("/user/signup", app.signupUser)
		r.Get("/user/login", app.loginUserForm)
		r.Post("/user/login", app.loginUser)

		r.Group(func(r chi.Router) {
			r.Use(app.requireAuthenticatedUser)

			r.Get("/snippets/create", app.createSnippetForm)
			r.Post("/snippets/create", app.createSnippet)
			r.Post("/user/logout", app.logoutUser)
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/healthcheck", app.healthCheckAPI)
		r.Get("/snippets", app.listSnippetsAPI)
		r.Get("/snippets/{id}", app.getSnippetAPI)
		r.Get("/tags", app.listTagsAPI)
		r.Get("/users/{id}", app.getUserAPI)

		r.Group(func(r chi.Router) {
			r.Use(app.requireAuthenticatedUser)
			r.Post("/snippets", app.createSnippetAPI)
		})
	})

	return r
}

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/newcastile/snippetbox/ui"
)

// The routes() method returns a servemux containing our application routes.
func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	// Update the pattern for the route for the static files.
	fileServer := http.FileServer(http.FS(ui.Files))

	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// And then create the routes using the appropriate methods, patterns and
	// handlers.

	// Protected routes
	router.Handler(http.MethodGet, "/snippet/create", app.sessionManager.LoadAndSave(noSurf(app.authenticate(app.requireAuthentication(http.HandlerFunc(app.snippetCreate))))))
	router.Handler(http.MethodPost, "/snippet/create", app.sessionManager.LoadAndSave(noSurf(app.authenticate(app.requireAuthentication(http.HandlerFunc(app.snippetCreatePost))))))
	router.Handler(http.MethodPost, "/user/logout", app.sessionManager.LoadAndSave(noSurf(app.authenticate(app.requireAuthentication(http.HandlerFunc(app.userLogoutPost))))))
	router.Handler(http.MethodPost, "/snippet/delete/:id", app.sessionManager.LoadAndSave(noSurf(app.authenticate(app.requireAuthentication(http.HandlerFunc(app.snippetDeletePost))))))

	// Unprotected routes
	router.Handler(http.MethodGet, "/", app.sessionManager.LoadAndSave(noSurf(app.authenticate(http.HandlerFunc(app.home)))))
	router.Handler(http.MethodGet, "/snippet/view/:id", app.sessionManager.LoadAndSave(noSurf(app.authenticate(http.HandlerFunc(app.snippetView)))))
	router.Handler(http.MethodGet, "/user/signup", app.sessionManager.LoadAndSave(noSurf(app.authenticate(http.HandlerFunc(app.userSignup)))))
	router.Handler(http.MethodPost, "/user/signup", app.sessionManager.LoadAndSave(noSurf(app.authenticate(http.HandlerFunc(app.userSignupPost)))))
	router.Handler(http.MethodGet, "/user/login", app.sessionManager.LoadAndSave(noSurf(app.authenticate(http.HandlerFunc(app.userLogin)))))
	router.Handler(http.MethodPost, "/user/login", app.sessionManager.LoadAndSave(noSurf(app.authenticate(http.HandlerFunc(app.userLoginPost)))))

	return app.recoverPanic(app.logRequest(secureHeaders(router)))
}

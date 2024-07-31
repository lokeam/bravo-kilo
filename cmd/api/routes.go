package main

import (
	"bravo-kilo/cmd/handlers"
	"bravo-kilo/cmd/middleware"
	"net/http"

	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes(h *handlers.Handlers) http.Handler {
	mux := chi.NewRouter()
	mux.Use(chimiddleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
    AllowedOrigins:   []string{"https://*", "http://*"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: true,
	}))

	mux.Get("/auth/google/signin", h.GoogleSignIn)
	mux.Get("/auth/google/callback", h.GoogleCallback)
	mux.Get("/auth/token/verify", h.VerifyToken)
	mux.Post("/auth/token/refresh", h.RefreshToken)
	mux.Post("/auth/signout", h.SignOut)

	mux.Route("/api/v1/user", func(r chi.Router) {
		r.Use(middleware.VerifyJWT)
		r.Get("/books", h.GetAllUserBooks)
		r.Get("/books/count", h.GetBooksCountByFormat)
	})

	mux.Route("/api/v1/books", func(r chi.Router) {
		r.Use(middleware.VerifyJWT)
		r.Get("/search", h.SearchBooks)
		r.Get("/{bookID}", h.GetBookByID)
		r.Put("/{bookID}", h.UpdateBook)
		r.Delete("/{bookID}", h.DeleteBook)
	})

	return mux
}

package main

import (
	"bravo-kilo/cmd/handlers"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes(h *handlers.Handlers) http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
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
	mux.Post("/auth/signout", h.SignOut)

	mux.Get("/api/v1/books/search", h.SearchBooks)
	mux.Get("/api/v1/user/books", h.GetAllUserBooks)
	mux.Get("/api/v1/books/{bookID}", h.GetBookByID)

	mux.Put("/api/v1/books/{bookID}", h.UpdateBook)

	return mux
}

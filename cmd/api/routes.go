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

	mux.Get("/auth/google/signin", h.HandleGoogleSignIn)
	mux.Get("/auth/google/callback", h.HandleGoogleCallback)
	mux.Get("/auth/token/verify", h.HandleVerifyToken)
	mux.Post("/auth/token/refresh", h.HandleRefreshToken)
	mux.Post("/auth/signout", h.HandleSignOut)

	mux.Route("/api/v1/user", func(r chi.Router) {
		r.Use(middleware.VerifyJWT)
		r.Get("/books", h.HandleGetAllUserBooks)
		r.Get("/books/authors", h.HandleGetBooksByAuthors)
		r.Get("/books/format", h.HandleGetBooksByFormat)
		r.Get("/books/genres", h.HandleGetBooksByGenres)

		// Apply rate limiting on uploads
		r.With(middleware.RateLimiter).Post("/upload", h.UploadCSV)
	})

	mux.Route("/api/v1/books", func(r chi.Router) {
		r.Use(middleware.VerifyJWT)
		r.Get("/by-id/{bookID}", h.HandleGetBookByID)
		r.Get("/search", h.HandleSearchBooks)
		r.Get("/by-title", h.HandleGetBookIDByTitle)
		r.Put("/{bookID}", h.HandleUpdateBook)
		r.Post("/add", h.HandleInsertBook)
		r.Delete("/{bookID}", h.HandleDeleteBook)
	})

	return mux
}

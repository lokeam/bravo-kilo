package main

import (
	"net/http"

	"github.com/lokeam/bravo-kilo/cmd/middleware"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	auth "github.com/lokeam/bravo-kilo/internal/shared/handlers/auth"

	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes(
	bookHandlers *handlers.BookHandlers,
	searchHandlers *handlers.SearchHandlers,
	authHandlers *auth.AuthHandlers,
	) http.Handler {
	mux := chi.NewRouter()

	mux.Use(chimiddleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
    AllowedOrigins:   []string{"https://*", "http://*"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: true,
	}))

	mux.Get("/auth/google/signin", authHandlers.HandleGoogleSignIn)
	mux.Get("/auth/google/callback", authHandlers.HandleGoogleCallback)
	mux.Get("/auth/token/verify", authHandlers.HandleVerifyToken)
	mux.Post("/auth/token/refresh", authHandlers.HandleRefreshToken)
	mux.Post("/auth/signout", authHandlers.HandleSignOut)

	mux.Route("/api/v1/user", func(r chi.Router) {
		r.Use(middleware.VerifyJWT)
		r.Get("/books", bookHandlers.HandleGetAllUserBooks)
		r.Get("/books/authors", bookHandlers.HandleGetBooksByAuthors)
		r.Get("/books/format", bookHandlers.HandleGetBooksByFormat)
		r.Get("/books/genres", bookHandlers.HandleGetBooksByGenres)
		r.Get("/books/homepage", bookHandlers.HandleGetHomepageData)

		// Apply rate limiting on uploads and exports"
		r.With(middleware.RateLimiter).Post("/upload", bookHandlers.UploadCSV)
		r.With(middleware.RateLimiter).Get("/export", bookHandlers.HandleExportUserBooks)
	})

	mux.Route("/api/v1/books", func(r chi.Router) {
		r.Use(middleware.VerifyJWT)
		r.Get("/by-id/{bookID}", bookHandlers.HandleGetBookByID)
		r.Get("/search", searchHandlers.HandleSearchBooks)
		r.With(middleware.RateLimiter).Get("/summary", bookHandlers.HandleGetGeminiBookSummary)
		r.Get("/by-title", bookHandlers.HandleGetBookIDByTitle)
		r.Put("/{bookID}", bookHandlers.HandleUpdateBook)
	r.Post("/add", bookHandlers.HandleInsertBook)
		r.Delete("/{bookID}", bookHandlers.HandleDeleteBook)
	})

	return mux
}

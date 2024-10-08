package main

import (
	"net/http"
	"os"

	"github.com/lokeam/bravo-kilo/cmd/middleware"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	auth "github.com/lokeam/bravo-kilo/internal/shared/handlers/auth"

	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/csrf"
)

var isProduction bool

func init() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	isProduction = env == "production"
}

func (app *application) routes(
	bookHandlers *handlers.BookHandlers,
	searchHandlers *handlers.SearchHandlers,
	authHandlers *auth.AuthHandlers,
) http.Handler {
	mux := chi.NewRouter()

	mux.Use(chimiddleware.Recoverer)
	mux.Use(middleware.LogHeaders)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	csrfMiddleware := csrf.Protect(
		[]byte(os.Getenv("CSRF_AUTH_KEY")),
		csrf.Secure(isProduction),
		csrf.HttpOnly(true),
		csrf.RequestHeader("X-CSRF-Token"),
		csrf.Path("/"),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log CSRF token mismatch details
			app.logger.Error("CSRF token mismatch",
				"method", r.Method,
				"url", r.URL.String(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
			http.Error(w, "Forbidden - CSRF token invalid", http.StatusForbidden)
		})),
	)

	mux.Group(func(r chi.Router) {
		r.Use(csrfMiddleware)
		r.Use(middleware.CSRFTokens)

		r.Route("/api/v1/user", func(r chi.Router) {
			r.Use(middleware.VerifyJWT)
			r.Get("/books", bookHandlers.HandleGetAllUserBooks)
			r.Get("/books/authors", bookHandlers.HandleGetBooksByAuthors)
			r.Get("/books/format", bookHandlers.HandleGetBooksByFormat)
			r.Get("/books/genres", bookHandlers.HandleGetBooksByGenres)
			r.Get("/books/homepage", bookHandlers.HandleGetHomepageData)
			r.Get("/books/tags", bookHandlers.HandleGetBooksByTags)
			// Apply rate limiting on uploads + exports
			r.With(middleware.RateLimiter).Post("/upload", bookHandlers.UploadCSV)
			r.With(middleware.RateLimiter).Get("/export", bookHandlers.HandleExportUserBooks)
		})

		r.Route("/api/v1/books", func(r chi.Router) {
			r.Use(middleware.VerifyJWT)
			r.Get("/by-id/{bookID}", bookHandlers.HandleGetBookByID)
			r.Get("/search", searchHandlers.HandleSearchBooks)
			r.With(middleware.RateLimiter).Get("/summary", bookHandlers.HandleGetGeminiBookSummary)
			r.Get("/by-title", bookHandlers.HandleGetBookIDByTitle)
			r.Put("/{bookID}", bookHandlers.HandleUpdateBook)
			r.Post("/add", bookHandlers.HandleInsertBook)
			r.Delete("/{bookID}", bookHandlers.HandleDeleteBook)
		})
	})

	// OAuth2 routes without CSRF protection
	mux.Group(func(r chi.Router) {
		mux.Get("/auth/google/signin", authHandlers.HandleGoogleSignIn)
		mux.Get("/auth/google/callback", authHandlers.HandleGoogleCallback)
		mux.Get("/auth/token/verify", authHandlers.HandleVerifyToken)
		mux.Post("/auth/token/refresh", authHandlers.HandleRefreshToken)
		mux.Post("/auth/signout", authHandlers.HandleSignOut)
		mux.Delete("/auth/delete-account", authHandlers.HandleDeleteAccount)
	})

	return mux
}

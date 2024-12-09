package main

import (
	"net/http"
	"os"
	"time"

	"github.com/lokeam/bravo-kilo/cmd/middleware"
	authhandlers "github.com/lokeam/bravo-kilo/internal/auth/handlers"
	"github.com/lokeam/bravo-kilo/internal/books/handlers"
	"github.com/lokeam/bravo-kilo/internal/shared/core"
	homehandlers "github.com/lokeam/bravo-kilo/internal/shared/home"
	libraryhandlers "github.com/lokeam/bravo-kilo/internal/shared/library"
	"github.com/lokeam/bravo-kilo/internal/shared/validator"

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
	authHandlers *authhandlers.AuthHandlers,
	libraryHandler *libraryhandlers.LibraryHandler,
	baseValidator *validator.BaseValidator,
	homeHandler *homehandlers.HomeHandler,
	) http.Handler {
	mux := chi.NewRouter()

	// Panic Recovery
	mux.Use(chimiddleware.Recoverer)

	// Ensure every request has a unique ID
	mux.Use(middleware.RequestID)

	// Logging
	mux.Use(middleware.LogHeaders)

	// CORS
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	// CSRF protection
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

		r.With(middleware.StandardRateLimiter).Get("/api/v1/csrf-token", authHandlers.HandleRefreshCSRFToken)

		r.Route("/api/v1/user", func(r chi.Router) {
			r.Use(middleware.VerifyJWT)
			r.Use(middleware.StandardRateLimiter)

			r.Get("/books", bookHandlers.HandleGetAllUserBooks)
			r.Get("/books/authors", bookHandlers.HandleGetBooksByAuthors)
			r.Get("/books/format", bookHandlers.HandleGetBooksByFormat)
			r.Get("/books/genres", bookHandlers.HandleGetBooksByGenres)
			r.Get("/books/homepage", bookHandlers.HandleGetHomepageData)
			r.Get("/books/tags", bookHandlers.HandleGetBooksByTags)

			// Apply intensive rate limiting on uploads + exports
			r.With(middleware.IntensiveRateLimiter).Post("/upload", bookHandlers.UploadCSV)
			r.With(middleware.IntensiveRateLimiter).Get("/export", bookHandlers.HandleExportUserBooks)
		})

		r.Route("/api/v1/books", func(r chi.Router) {
			r.Use(middleware.VerifyJWT)
			r.Use(middleware.RequestValidation(baseValidator, middleware.ValidationConfig{
				Domain: core.BookDomainType,
				Timeout: 30 * time.Second,
			}))

			// Standard rate limiting for bookID
			r.With(middleware.StandardRateLimiter).Get("/by-id/{bookID}", bookHandlers.HandleGetBookByID)

			// More restrictive rate limiting for search
			r.With(middleware.IntensiveRateLimiter).Get("/search", searchHandlers.HandleSearchBooks)

			// Standard rate limiting for summary + bookID
			r.With(middleware.StandardRateLimiter).Get("/summary", bookHandlers.HandleGetGeminiBookSummary)
			r.With(middleware.StandardRateLimiter).Get("/by-title", bookHandlers.HandleGetBookIDByTitle)

			r.With(middleware.StandardRateLimiter).Put("/{bookID}", bookHandlers.HandleUpdateBook)
			r.With(middleware.StandardRateLimiter).Post("/add", bookHandlers.HandleInsertBook)
			r.With(middleware.StandardRateLimiter).Delete("/{bookID}", bookHandlers.HandleDeleteBook)
		})

		r.Route("/api/v1/pages", func(r chi.Router) {
			r.Use(middleware.VerifyJWT)
			r.Use(middleware.RequestValidation(baseValidator, middleware.ValidationConfig{
				Domain: core.BookDomainType,
				Timeout: 30 * time.Second,
			}))

			r.Use(middleware.NewAdaptiveCompression(app.compressionMonitor))
			r.Use(middleware.StandardRateLimiter)

			r.Get("/library", libraryHandler.HandleGetLibraryPageData)
			r.Get("/home", homeHandler.HandleGetHomePageData)
		})
	})

	// OAuth2 routes without CSRF protection
	mux.Group(func(r chi.Router) {
		r.With(middleware.CriticalRateLimiter).Get("/auth/google/signin", authHandlers.HandleGoogleSignIn)
		r.With(middleware.CriticalRateLimiter).Get("/auth/google/callback", authHandlers.HandleGoogleCallback)
		r.With(middleware.CriticalRateLimiter).Get("/auth/token/verify", authHandlers.HandleVerifyToken)
		r.With(middleware.CriticalRateLimiter).Post("/auth/token/refresh", authHandlers.HandleRefreshToken)
		r.With(middleware.CriticalRateLimiter).Post("/auth/signout", authHandlers.HandleSignOut)
		r.With(middleware.CriticalRateLimiter).Delete("/auth/delete-account", authHandlers.HandleDeleteAccount)
	})

	return mux
}

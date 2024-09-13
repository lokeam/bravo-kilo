package services

import (
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/books/services"
	// Import other services when available
)

type Services struct {
    BookService  services.BookService
		BookRepository repository.BookRepository
    // GameService  GameService  // Placeholder for future domain services
    // MovieService MovieService // Placeholder for future domain services
}

// Helper methods initializing future services

package authservice

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/lokeam/bravo-kilo/config"
	"github.com/lokeam/bravo-kilo/internal/books/repository"
	"github.com/lokeam/bravo-kilo/internal/shared/models"
	"github.com/lokeam/bravo-kilo/internal/shared/transaction"
)

type AuthService interface {
	DeleteUser(ctx context.Context, userID int) error
	MarkUserForDeletion(ctx context.Context, userID int) error
	ProcessDeletionQueue(ctx context.Context) error
	ProcessAccountDeletion(ctx context.Context, userID int) error
}

type AuthServiceImpl struct {
	userRepo models.UserRepository
	tokenRepo models.TokenModel
	bookRedisCache repository.BookRedisCache
	dbManager transaction.DBManager
	models models.Models
	logger *slog.Logger
}

func NewAuthService(
	userRepo models.UserRepository,
	tokenRepo models.TokenModel,
	bookRedisCache repository.BookRedisCache,
	dbManager transaction.DBManager,
	logger *slog.Logger,
) AuthService {
	return &AuthServiceImpl{
		userRepo: userRepo,
		tokenRepo: tokenRepo,
		bookRedisCache: bookRedisCache,
		dbManager: dbManager,
		logger: logger,
	}
}

func (s *AuthServiceImpl) MarkUserForDeletion(ctx context.Context, userID int) error {
	s.logger.Info("Marking user for deletion", "userID", userID)

	tx, err := s.dbManager.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("Error beginning transaction", "error", err)
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer s.dbManager.RollbackTransaction(tx)

	deletionTime := time.Now()
	err = s.userRepo.MarkForDeletion(ctx, tx, userID, deletionTime)
	if err != nil {
		return fmt.Errorf("error marking user for deletion: %w", err)
	}

	err = s.tokenRepo.DeleteByUserID(userID)
	if err != nil {
		return fmt.Errorf("error deleting tokens: %w", err)
	}

	err = s.bookRedisCache.SetUserDeletionMarker(ctx, strconv.Itoa(userID), config.AppConfig.UserDeletionMarkerExpiration)
	if err != nil {
		s.logger.Error("Error setting user deletion marker in Redis", "error", err)
	}

	s.logger.Info("User successfully marked for deletion", "userID", userID)
	return nil
}

func (s *AuthServiceImpl) ProcessDeletionQueue(ctx context.Context) error {
	userIDs, err := s.bookRedisCache.GetDeletionQueue(ctx)
	if err != nil {
		return fmt.Errorf("error getting deletion queue: %w", err)
	}

	for _, userIDStr := range userIDs {
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			s.logger.Error("Error converting userID to int", "userID", userID, "error", err)
			continue
		}

		err = s.bookRedisCache.RemoveFromDeletionQueue(ctx, userIDStr)
		if err != nil {
			s.logger.Error("Error removing user from deletion queue", "userID", userIDStr, "error", err)
		}
	}

	return nil
}

func (s *AuthServiceImpl) DeleteUser(ctx context.Context, userID int) error {
	// Start a transaction
	tx, err := s.dbManager.BeginTransaction(ctx)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer s.dbManager.RollbackTransaction(tx)


	// Get a list of all bookIDs belonging to user
	bookIDs, err := s.userRepo.GetUserBookIDs(userID)
	if err != nil {
		return fmt.Errorf("error getting user's book IDs: %w", err)
	}

	// Loop through list of bookIds and use the book_deleter to delete each book and its association
	bookDeleter, err := repository.NewBookDeleter(s.dbManager.GetDB(), s.logger)
	if err != nil {
		return fmt.Errorf("error creating book deleter: %w", err)
	}

	for _, bookID := range bookIDs {
		err = bookDeleter.Delete(bookID)
		if err != nil {
			return fmt.Errorf("error deleting book %d: %w", bookID, err)
		}
	}

	// Delete all tokens associated with user
	err = s.models.Token.DeleteByUserID(userID)
	if err != nil {
		return fmt.Errorf("error deleting user tokens: %w", err)
	}


	// Delete user from users table
	err = s.userRepo.Delete(userID)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	// Commit transaction
	err = s.dbManager.CommitTransaction(tx)
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (s *AuthServiceImpl) ProcessAccountDeletion(ctx context.Context, userID int) error {
	s.logger.Info("Handling account deletion", "userID", userID)

	// Begin transaction
	tx, err := s.dbManager.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error("Error beginning transaction", "error", err)
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer s.dbManager.RollbackTransaction(tx)

	// Soft delete (mark for deletion)
	deletionTime := time.Now()
	err = s.userRepo.MarkForDeletion(ctx, tx, userID, deletionTime)
	if err != nil {
		s.logger.Error("Error marking user for deletion", "error", err)
		return fmt.Errorf("error marking user for deletion: %w", err)
	}
	s.logger.Info("User successfully marked for deletion", "userID", userID)

	// Delete all refresh tokens
	err = s.tokenRepo.DeleteByUserID(userID)
	if err != nil {
		s.logger.Error("Error deletion refresh tokens", "error", err)
		return fmt.Errorf("error deleting refresh tokens: %w", err)
	}
	s.logger.Info("Refresh tokens successfully deleted", "userID", userID)

	// Commit transaction
	err = s.dbManager.CommitTransaction(tx)
	if err != nil {
		s.logger.Error("Error committing transaction", "error", err)
		return fmt.Errorf("error processing request: %w", err)
	}
	s.logger.Info("Transaction comitted successfully")

	// Set deletion marker in Redis
	err = s.bookRedisCache.SetUserDeletionMarker(ctx, strconv.Itoa(userID), config.AppConfig.UserDeletionMarkerExpiration)
	if err != nil {
		s.logger.Error("Error setting user deletion marker in Redis", "error", err)
	}
	s.logger.Info("User deletion marker set in Redis", "userID", userID)


	// Add user to deletion queue in Redis
	err = s.bookRedisCache.AddToDeletionQueue(ctx, strconv.Itoa(userID))
	if err != nil {
		s.logger.Error("Error adding user to deletion queue in Redis", "error", err)
	}
	s.logger.Info("User added to deletion queue in Redis", "userID", userID)

	return nil
}

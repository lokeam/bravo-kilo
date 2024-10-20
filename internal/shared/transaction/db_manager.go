package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

// DBManager interface handles transaction lifecycle and DB-level utilities.
type DBManager interface {
    BeginTransaction(ctx context.Context) (*sql.Tx, error)
    CommitTransaction(tx *sql.Tx) error
    RollbackTransaction(tx *sql.Tx) error
    GetDB() *sql.DB
}

// DBManagerImpl is the concrete implementation.
type DBManagerImpl struct {
    DB     *sql.DB
    Logger *slog.Logger
}

func NewDBManager(db *sql.DB, logger *slog.Logger) (DBManager, error) {
    if db == nil || logger == nil {
        return nil, fmt.Errorf("database or logger cannot be nil")
    }

    return &DBManagerImpl{
        DB:     db,
        Logger: logger,
    }, nil
}

// Pass Actual database connection
func (d *DBManagerImpl) GetDB() *sql.DB {
	return d.DB
}

// BeginTransaction starts a new database transaction.
func (d *DBManagerImpl) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
    tx, err := d.DB.BeginTx(ctx, nil)
    if err != nil {
        d.Logger.Error("Error beginning transaction", "error", err)
        return nil, err
    }
    return tx, nil
}

// CommitTransaction commits a transaction.
func (d *DBManagerImpl) CommitTransaction(tx *sql.Tx) error {
    if err := tx.Commit(); err != nil {
        d.Logger.Error("Error committing transaction", "error", err)
        return err
    }
    return nil
}

// RollbackTransaction rolls back a transaction in case of an error.
func (d *DBManagerImpl) RollbackTransaction(tx *sql.Tx) error {
    if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
        d.Logger.Error("Error rolling back transaction", "error", err)
        return err
    }
    return nil
}

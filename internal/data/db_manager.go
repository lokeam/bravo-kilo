package data

import (
	"context"
	"database/sql"
	"log/slog"
)

// DBManager interface handles transaction lifecycle and DB-level utilities.
type DBManager interface {
    BeginTransaction(ctx context.Context) (*sql.Tx, error)
    CommitTransaction(tx *sql.Tx) error
    RollbackTransaction(tx *sql.Tx) error
}

type DBManagerImpl struct {
    DB     *sql.DB
    Logger *slog.Logger
}


func (d *DBManagerImpl) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
			d.Logger.Error("Error beginning transaction", "error", err)
			return nil, err
	}
	return tx, nil
}

func (d *DBManagerImpl) CommitTransaction(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
			d.Logger.Error("Error committing transaction", "error", err)
			return err
	}
	return nil
}

func (d *DBManagerImpl) RollbackTransaction(tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			d.Logger.Error("Error rolling back transaction", "error", err)
			return err
	}
	return nil
}

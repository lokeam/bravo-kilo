package driver

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type DB struct {
	SQL *sql.DB
}

var dbConnection = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute

func ConnectPostgres(dsn string, logger *slog.Logger) (*DB, error) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("Error opening database", "error", err)
		return nil, err
	}

	database.SetMaxOpenConns(maxOpenDbConn)
	database.SetMaxIdleConns(maxIdleDbConn)
	database.SetConnMaxLifetime(maxDbLifetime)

	err = testDB(database, logger)
	if err != nil {
		logger.Error("Error testing db connection", "error", err)
		return nil, err
	}

	dbConnection.SQL = database
	return dbConnection, nil
}

func testDB(database *sql.DB, logger *slog.Logger) error {
	err := database.Ping()
	if err != nil {
		logger.Error("DB ping failed ", "error", err)
		return err
	}
	logger.Info("***** DB ping successful *****")

	return nil
}

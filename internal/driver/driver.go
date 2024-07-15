package driver

import (
	"database/sql"
	"fmt"
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

func ConnectPostgres(dsn string) (*DB, error) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(maxOpenDbConn)
	database.SetMaxIdleConns(maxIdleDbConn)
	database.SetConnMaxLifetime(maxDbLifetime)

	err = testDB(database)
	if err != nil {
		return nil, err
	}

	dbConnection.SQL = database
	return dbConnection, nil
}

func testDB(database *sql.DB) error {
	err := database.Ping()
	if err != nil {
		fmt.Println("Error testing database! ", err)
		return err
	}
	fmt.Println("***** DB ping successful *****")

	return nil
}

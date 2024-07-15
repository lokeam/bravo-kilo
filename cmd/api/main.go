package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bravo-kilo/cmd/api/handlers"
	"bravo-kilo/config"
	"bravo-kilo/internal/driver"

	"github.com/joho/godotenv"
)

type appConfig struct {
	port int
}

type application struct {
	config    appConfig
	errorLog  *log.Logger
	infoLog   *log.Logger
	db        *driver.DB
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	var cfg appConfig
	cfg.port = 8081

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	dataSrcName := "host=localhost port=7654 user=postgres password=password dbname=bkapi sslmode=disable timezone=UTC connect_timeout=5"
	fmt.Println("-----------------------------")
	fmt.Println("DSN:", dataSrcName) // Print DSN for verification
	fmt.Println("DSN from environment:", os.Getenv("DSN"))

	db, err := driver.ConnectPostgres(dataSrcName)
	if err != nil {
		errorLog.Fatal("Cannot connect to database")
	}

	defer db.SQL.Close()

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		db: db,
	}

	// Initialize the config package with the infoLog
	config.InitConfig(infoLog)

	// Initialize handlers with the infoLog
	h := handlers.NewHandlers(infoLog)

	err = app.serve(h)
	if err != nil {
		log.Fatal(err)
	}
}

func (app *application) serve(h *handlers.Handlers) error {
	app.infoLog.Println("API listening on port", app.config.port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(h),
	}

	return srv.ListenAndServe()
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bravo-kilo/cmd/api/handlers"
	"bravo-kilo/config"

	"github.com/joho/godotenv"
)

type appConfig struct {
	port int
}

type application struct {
	config   appConfig
	errorLog *log.Logger
	infoLog  *log.Logger
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

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
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

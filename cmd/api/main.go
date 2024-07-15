package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"bravo-kilo/config"
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
	var cfg appConfig
	cfg.port = 8081

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
	}

	config.InitConfig(infoLog)

	err := app.serve()
	if err != nil {
		log.Fatal(err)
	}
}

func (app *application) serve() error {
	app.infoLog.Println("API listening on port", app.config.port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
	}

	return srv.ListenAndServe()
}

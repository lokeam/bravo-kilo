package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type config struct {
	port int
}

type application struct {
	config config
	errorLog *log.Logger
	infoLog *log.Logger
}

func main () {
	var cfg config
	cfg.port = 8081

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		config: cfg,
		infoLog: infoLog,
		errorLog: errorLog,
	}

	err := app.serve()
	if err != nil {
		log.Fatal(err)
	}
}

func (app *application) serve() error {
	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		var payload struct {
			GolangBackend bool `json:"golangBackend"`
			Message string `json:"message"`
		}
		payload.GolangBackend = true
		payload.Message = "Hello, world"

		out, err := json.MarshalIndent(payload, "", "\t")
		if err != nil {
			app.errorLog.Println(err)
		}

		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)
		response.Write(out)
	})

	app.infoLog.Println("API listening on port", app.config.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", app.config.port), nil)
}
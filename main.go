package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/youstinus/email-sender/domains"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type App struct {
	db *mongo.Database
}

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("mongodb_url")))
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Connect(context.TODO()); err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database(os.Getenv("mongodb_database"))

	app := App{db}

	r := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true,
	})
	// use swagger UI if you have to.
	//fs := http.FileServer(http.Dir("swagger-ui/dist/"))
	//r.PathPrefix("/").Handler(fs)

	content := r.PathPrefix("/v1/").Subrouter()
	content.HandleFunc("/emails", app.GetAllEmails).Methods("GET")
	content.HandleFunc("/emails", app.CreateEmail).Methods("POST")
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)
	handler := c.Handler(loggedRouter)
	log.Fatal(http.ListenAndServe(":8000", handler))
}

func toJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(&data); err != nil {
		log.Fatal(err)
	}
}

func fromJSON(r io.Reader, dest interface{}) {
	if err := json.NewDecoder(r).Decode(&dest); err != nil {
		log.Fatal(err)
	}
}

func (app *App) GetAllEmails(w http.ResponseWriter, r *http.Request) {
	contents, err := domains.GetAllEmails(app.db)
	if err != nil {
		log.Fatal(err)
	}
	toJSON(w, contents)
}

func (app *App) CreateEmail(w http.ResponseWriter, r *http.Request) {
	var c domains.Email
	fromJSON(r.Body, &c)
	rc, err := domains.CreateEmail(app.db, c)
	if err != nil {
		log.Fatal(err)
	}
	toJSON(w, rc)
}

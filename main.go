package main

import (
	"log"
	"net/http"
	"os"

	"gitlab.com/pagalguy/loki/database"
	"gitlab.com/pagalguy/loki/handlers"
	"google.golang.org/appengine"
)

func main() {

	databaseURL := os.Getenv("POSTGRES_CONNECTION")

	// Database connection
	db := database.InitDB(databaseURL)

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping to database: %v\n", err)
	}

	router := handlers.CreateRouter(db)

	// go handlers.StartPubsubListen(db)

	http.Handle("/", router)
	appengine.Main()
}

package main

import (
	"log"
	"net/http"
	"os"

	"google.golang.org/appengine"

	"gitlab.com/pagalguy/loki/database"
	"gitlab.com/pagalguy/loki/handlers"
)

func main() {

	databaseURL := os.Getenv("POSTGRES_CONNECTION")

	// Database connection
	Db := database.InitDB(databaseURL)

	defer Db.Close()

	if err := Db.Ping(); err != nil {
		log.Fatalf("could not ping to database: %v\n", err)
	}

	router := handlers.CreateRouter(Db)

	http.Handle("/", router)
	appengine.Main()
}

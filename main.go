package main

import (
	"log"
	"net/http"
	"os"

	"gitlab.com/pagalguy/loki/database"
	"gitlab.com/pagalguy/loki/handlers"
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

	go handlers.StartPubsubListen(db)

	// API Server
	port := env("PORT", "8080")
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))

}

func env(key, fallbackValue string) string {

	value, isPresent := os.LookupEnv(key)

	if !isPresent {
		return fallbackValue
	}

	return value
}

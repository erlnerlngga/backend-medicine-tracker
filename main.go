package main

import (
	"log"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	store, err := NewPostgresStore();

	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	defer store.db.Close()
	server := NewApiServer(":8080", store)
	server.Run()
}

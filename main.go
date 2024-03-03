package main

import (
	"Chat/db"
	"log"
	"net/http"
	"os"
	"Chat/handlers"
	"github.com/joho/godotenv"
)


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	pool := db.DbStart(databaseURL)

	db := db.NewDB(pool)
	handler := handlers.NewBaseHandler(db)
	http.HandleFunc("/ws", handler.HandleConnections)

	log.Println("Server started on :8000")
	er := http.ListenAndServe(":8000", nil)
	if er != nil {
		log.Fatal("Error starting server: ", err)
	}
}

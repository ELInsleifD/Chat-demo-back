package main

import (
	"chat-demo/api"
	"chat-demo/connect"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env file")
	}
}
func main() {
	port := os.Getenv("PORT")
	dsnStr := os.Getenv("MONGO_DSN")

	db, err := connect.Connection(dsnStr)
	if err != nil {
		panic(fmt.Sprintf("Error connecting to db: %s", err))
	}

	srv, err := api.Create(db)
	if err != nil {
		panic(fmt.Sprintf("Error creating server: %s", err))
	}
	panic(srv.Listen(port))

}

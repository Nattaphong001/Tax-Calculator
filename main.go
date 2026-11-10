package main

import (
	"log"
	"net/http"
	"os"

	"tax-calculator/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler.NewRouter()))
}
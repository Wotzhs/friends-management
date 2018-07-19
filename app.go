package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := ":3000"
	if os.Getenv("GO_ENV") == "test" {
		port = ":3001"
	}

	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

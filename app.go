package main

import (
	"log"
	"net/http"
)

func main() {
	server := &http.Server{
		Addr:    ":3000",
		Handler: router,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

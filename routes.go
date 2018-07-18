package main

import (
	"encoding/json"
	"net/http"
)

var mux *http.ServeMux

func init() {
	mux = http.NewServeMux()
	mux.HandleFunc("/", helloWorld)
	mux.HandleFunc("/api/", helloWorldApi)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

type helloWorldType struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func helloWorldApi(w http.ResponseWriter, r *http.Request) {
	res, err := json.Marshal(helloWorldType{"hello world", true})
	if err != nil {
		w.Write([]byte("error in json marshal"))
	}
	w.Write(res)
}

package main

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var router *httprouter.Router
var jsonRes []byte

type apiResponse struct {
	Success bool   `json:"success"`
	Errors  string `json:"errors,omitempty"`
}

func init() {
	router = httprouter.New()
	router.POST("/api/friends", CreateFriendsHandler)
}

func CreateFriendsHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		jsonRes, _ = json.Marshal(&apiResponse{Success: false, Errors: err.Error()})
		w.Write(jsonRes)
		return
	}

	users := []string{}

	if err := json.Unmarshal([]byte(r.FormValue("friends")), &users); err != nil {
		w.Write([]byte(`{"status": false}`))
	}

	if err := createFriends(users); err != nil {
		jsonRes, _ = json.Marshal(&apiResponse{Success: false, Errors: err.Error()})
		w.Write(jsonRes)
		return
	}

	jsonRes, _ = json.Marshal(&apiResponse{Success: true})
	w.Write(jsonRes)
}

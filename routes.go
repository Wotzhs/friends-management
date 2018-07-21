package main

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var router *httprouter.Router
var jsonRes []byte

type apiResponse struct {
	Status bool  `json:"status"`
	Errors error `json:"errors,omitempty"`
}

func init() {
	router = httprouter.New()
	router.POST("/api/friends", CreateFriendsHandler)
}

func CreateFriendsHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		jsonRes, _ = json.Marshal(&apiResponse{Status: false, Errors: err})
		w.Write(jsonRes)
	}

	users := []string{}

	if err := json.Unmarshal([]byte(r.FormValue("friends")), &users); err != nil {
		w.Write([]byte(`{"status": false}`))
	}

	if err := createFriends(users); err != nil {
		jsonRes, _ = json.Marshal(&apiResponse{Status: false, Errors: err})
		w.Write(jsonRes)
	}

	jsonRes, _ = json.Marshal(&apiResponse{Status: true})
	w.Write(jsonRes)
}

package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var router *httprouter.Router
var jsonRes []byte

type apiResponse struct {
	Success bool     `json:"success"`
	Errors  string   `json:"errors,omitempty"`
	Friends []string `json:"friends,omitempty"`
	Count   int      `json:"count,omitempty"`
}

func init() {
	router = httprouter.New()
	router.POST("/api/friends", createFriendsHandler)
	router.GET("/api/friends", getFriendsListHandler)

}

func createFriendsHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if err := r.ParseForm(); err != nil {
		jsonRes, _ = json.Marshal(&apiResponse{Success: false, Errors: err.Error()})
		w.Write(jsonRes)
		return
	}

	friendsData := r.FormValue("friends")
	if friendsData == "" {
		jsonRes, _ = json.Marshal(&apiResponse{Success: false, Errors: "no users were provided"})
		w.Write(jsonRes)
		return
	}

	users := []string{}
	if err := json.Unmarshal([]byte(friendsData), &users); err != nil {
		jsonRes, _ = json.Marshal(&apiResponse{Success: false, Errors: err.Error()})
		w.Write(jsonRes)
	}

	if err := createFriends(users); err != nil {
		jsonRes, _ = json.Marshal(&apiResponse{Success: false, Errors: err.Error()})
		w.Write(jsonRes)
		return
	}

	jsonRes, _ = json.Marshal(&apiResponse{Success: true})
	w.Write(jsonRes)
}

func getFriendsListHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	a := struct {
		Email string
	}{}
	if err := json.Unmarshal(bodyBytes, &a); err != nil {
		jsonRes, _ = json.Marshal(&apiResponse{Success: false, Errors: err.Error()})
		w.Write(jsonRes)
		return
	}
	if a.Email == "" {
		jsonRes, _ = json.Marshal(&apiResponse{Success: false, Errors: "no email was provided"})
		w.Write(jsonRes)
		return
	}

	friendsList, count, err := getFriendsList(a.Email)
	if err != nil {
		jsonRes, _ = json.Marshal(&apiResponse{Success: false, Errors: err.Error(), Friends: friendsList, Count: count})
		w.Write(jsonRes)
		return
	}

	jsonRes, _ = json.Marshal(&apiResponse{Success: true, Errors: "", Friends: friendsList, Count: count})
	w.Write(jsonRes)
}

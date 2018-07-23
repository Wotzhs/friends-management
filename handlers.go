package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func createFriendsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	friends := &user{}
	if err := json.Unmarshal(bodyBytes, friends); err != nil {
		w.Write(makeSimpleResponse(fmt.Sprintf("invalid data err: %v", err)))
		return
	}

	err := friends.createFriends()
	w.Write(makeNewResponse(friends, err))
}

func getFriendsListHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	user := &user{}
	if err := json.Unmarshal(bodyBytes, &user); err != nil {
		w.Write(makeSimpleResponse(fmt.Sprintf("invalid data err: %v", err)))
		return
	}

	err := user.getFriends()
	w.Write(makeNewResponse(user, err))
}

func getCommonFriendsListHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	friends := &user{}
	if err := json.Unmarshal(bodyBytes, &friends); err != nil {
		w.Write(makeSimpleResponse(fmt.Sprintf("invalid data err: %v", err)))
		return
	}

	err := friends.getCommonFriends()
	w.Write(makeNewResponse(friends, err))
}

func subscribeUpdatesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	userRequest := &userRequest{}
	if err := json.Unmarshal(bodyBytes, &userRequest); err != nil {
		w.Write(makeSimpleResponse(fmt.Sprintf("invalid data err: %v", err)))
		return
	}

	err := userRequest.subscribeUpdates()
	if err != nil {
		w.Write(makeSimpleResponse(err.Error()))
		return
	}

	w.Write(makeSimpleResponse(""))
}

func blockUpdatesHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	userRequest := &userRequest{}
	if err := json.Unmarshal(bodyBytes, &userRequest); err != nil {
		w.Write(makeSimpleResponse(fmt.Sprintf("invalid data err: %v", err)))
		return
	}

	err := userRequest.blockUpdates()
	if err != nil {
		w.Write(makeSimpleResponse(err.Error()))
		return
	}

	w.Write(makeSimpleResponse(""))
}

func getSubscribedListHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	message := message{}
	if err := json.Unmarshal(bodyBytes, &message); err != nil {
		w.Write(makeSimpleResponse(fmt.Sprintf("invalid data err: %v", err)))
		return
	}

	user, err := message.getSubscribers()
	w.Write(makeNewResponse(&user, err))
}

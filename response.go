package main

import (
	"encoding/json"
	"log"
)

type handlerResponse struct {
	Success    bool     `json:"success"`
	Errors     string   `json:"errors,omitempty"`
	Friends    []string `json:"friends,omitempty"`
	Count      int      `json:"count,omitempty"`
	Recipients []string `json:"recipients,omitempty"`
}

type response interface {
	listFriends() []string
	getCount() int
	listSubscribers() []string
}

func makeNewResponse(r response, err error) json.RawMessage {
	success := true
	var errString string
	if err != nil {
		success = false
		errString = err.Error()
	}
	res := handlerResponse{
		Success:    success,
		Errors:     errString,
		Friends:    r.listFriends(),
		Count:      r.getCount(),
		Recipients: r.listSubscribers(),
	}
	json, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
	}
	return json
}

func makeSimpleResponse(errMessage string) json.RawMessage {
	handlerResponse := &handlerResponse{Success: true}
	if errMessage != "" {
		handlerResponse.Success = false
		handlerResponse.Errors = errMessage
	}
	json, err := json.Marshal(handlerResponse)
	if err != nil {
		log.Println(err)
	}
	return json
}

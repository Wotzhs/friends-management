package main

import (
	"github.com/julienschmidt/httprouter"
)

var router *httprouter.Router

func init() {
	router = httprouter.New()
	router.POST("/api/friends", createFriendsHandler)
	router.GET("/api/friends", getFriendsListHandler)
	router.GET("/api/friends/common", getCommonFriendsListHandler)
	router.POST("/api/friends/subscribe", subscribeUpdatesHandler)
	router.POST("/api/friends/block", blockUpdatesHandler)
	router.GET("/api/friends/subscribe", getSubscribedListHandler)
}

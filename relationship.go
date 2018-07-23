package main

import (
	"errors"
	"strings"
)

type relationship struct {
	Requestor string
	Target    string
	Status    string
}

type relationships []relationship

func (r relationships) isBlocked() (isBlocked bool, err error) {
	messages := []string{}
	for _, relationship := range r {
		if relationship.Status == relationshipIsBlocked {
			messages = append(messages, relationship.Requestor+" has blocked "+relationship.Target)
			isBlocked = true
		}
	}
	return isBlocked, errors.New(strings.Join(messages, ","))
}

func (r relationships) isFriend() (isFriend bool, err error) {
	messages := []string{}
	for _, relationship := range r {
		if relationship.Status == relationshipIsFriend {
			messages = append(messages, relationship.Requestor+" is already a friend of "+relationship.Target)
			isFriend = true
		}
	}
	return isFriend, errors.New(strings.Join(messages, ","))
}

func (r relationships) isSubscribed() (isSubscribed bool, err error) {
	messages := []string{}
	for _, relationship := range r {
		if relationship.Status == relationshipIsSubscribed {
			messages = append(messages, relationship.Requestor+" has already subscribed to "+relationship.Target)
			isSubscribed = true
		}
	}
	return isSubscribed, errors.New(strings.Join(messages, ","))
}

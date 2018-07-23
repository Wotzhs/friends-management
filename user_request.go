package main

import (
	"errors"
	"strings"
)

type userRequest struct {
	Requestor string
	Target    string
}

func (u userRequest) subscribeUpdates() error {
	if u.Requestor == "" {
		return errors.New("no requestor was provided")
	}
	if u.Target == "" {
		return errors.New("no target was provided")
	}

	requestor := strings.ToLower(u.Requestor)
	target := strings.ToLower(u.Target)

	users := []string{requestor, target}
	exists, relationships, err := ifExistsRelationship(users)
	if err != nil {
		return err
	}

	if exists {
		if isBlocked, err := relationships.isBlocked(); isBlocked {
			return err
		}
		if isFriend, err := relationships.isFriend(); isFriend {
			return err
		}
		if isSubscribed, err := relationships.isSubscribed(); isSubscribed {
			return err
		}
	}
	return subscribeUpdates(requestor, target)
}

func (u userRequest) blockUpdates() error {
	if u.Requestor == "" {
		return errors.New("no requestor was provided")
	}
	if u.Target == "" {
		return errors.New("no target was provided")
	}

	requestor := strings.ToLower(u.Requestor)
	target := strings.ToLower(u.Target)
	users := []string{requestor, target}
	exists, relationships, err := ifExistsRelationship(users)
	if err != nil {
		return err
	}

	if exists {
		if isBlocked, err := relationships.isBlocked(); isBlocked {
			return err
		}
		if isFriend, _ := relationships.isFriend(); isFriend {
			return blockExistingRelationship(u.Requestor, u.Target)
		}
		if isSubscribed, _ := relationships.isSubscribed(); isSubscribed {
			return blockExistingRelationship(u.Requestor, u.Target)
		}
	}

	return blockUpdates(u.Requestor, u.Target)
}

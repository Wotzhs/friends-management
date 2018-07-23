package main

import (
	"errors"
)

type user struct {
	Email       string
	Friends     []string
	QueryStatus bool
	Subscribers []string
}

func (u *user) createFriends() error {
	if len(u.Friends) != 2 {
		return errors.New("incorrect number of friends")
	}

	for _, user := range u.Friends {
		if !isEmailValid(user) {
			return errors.New("invalid email being submitted")
		}
	}

	if u.Friends[0] == u.Friends[1] {
		return errors.New("cannot be friends with oneself")
	}

	exists, relationships, err := ifExistsRelationship(u.Friends)
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

	return createFriends(u.Friends)
}

func (u *user) getFriends() error {
	if !isEmailValid(u.Email) {
		return errors.New("invalid user")
	}
	friends, err := getFriendsList(u.Email)
	if err != nil {
		return err
	}
	u.Friends = friends
	return nil
}

func (u *user) getCommonFriends() error {
	if len(u.Friends) != 2 {
		return errors.New("incorrect number of friends")
	}

	for _, user := range u.Friends {
		if !isEmailValid(user) {
			return errors.New("invalid user")
		}
	}

	exists, relationships, err := ifExistsRelationship(u.Friends)
	if err != nil {
		return err
	}

	if exists {
		if isBlocked, err := relationships.isBlocked(); isBlocked {
			return err
		}
	}

	friends, err := getCommonFriendsList(u.Friends)
	if err != nil {
		return err
	}

	u.Friends = friends
	return nil
}

func (u *user) getSubscribers() {

}

func (u *user) getQueryStatus() bool {
	return u.QueryStatus
}

func (u *user) listFriends() []string {
	return u.Friends
}

func (u *user) getCount() int {
	return len(u.Friends)
}

func (u *user) listSubscribers() []string {
	return u.Subscribers
}

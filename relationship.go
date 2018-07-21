package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type relationship struct {
	requestor string
	target    string
	status    string
}

const (
	relationshipIsFriend     = "friend"
	relationshipIsBlocked    = "blocked"
	relationshipIsSubscribed = "subscribed"
)

type relationships []relationship

func createFriends(users []string) error {
	if len(users) != 2 {
		return errors.New("incorrect number of friends")
	}

	if users[0] == users[1] {
		return errors.New("cannot be friends with oneself")
	}

	exists, relationships, err := ifExistsRelationship(users)
	if err != nil {
		return err
	}

	if exists {
		if isBlocked, message := isBlocked(relationships); isBlocked {
			return errors.New(message)
		}
		if isFriend, message := isFriend(relationships); isFriend {
			return errors.New(message)
		}
	}

	insertQuery := `
		INSERT INTO relationships (requestor, target, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	if _, err := db.Exec(insertQuery, users[0], users[1], relationshipIsFriend, now, now); err != nil {
		return err
	}

	now = time.Now()
	if _, err := db.Exec(insertQuery, users[1], users[0], relationshipIsFriend, now, now); err != nil {
		return err
	}
	return nil
}

func ifExistsRelationship(users []string) (exists bool, relationships relationships, err error) {
	statusQuery := `
		SELECT requestor, target, status FROM relationships 
		WHERE (requestor=$1 AND target=$2)
		OR (requestor=$2 AND target=$1)
	`

	rows, err := db.Query(statusQuery, users[0], users[1])
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to check if any relationships exists between the users %v", err))
		return
	}

	defer rows.Close()

	for rows.Next() {
		row := relationship{}
		err = rows.Scan(&row.requestor, &row.target, &row.status)
		if err != nil {
			return
		}
		relationships = append(relationships, row)
	}

	if len(relationships) == 0 {
		return
	}

	exists = true
	return
}

func isBlocked(relationships relationships) (bool, string) {
	messages := []string{}
	isBlocked := false
	for _, relationship := range relationships {
		if relationship.status == relationshipIsBlocked {
			messages = append(messages, relationship.requestor+" has blocked "+relationship.target)
			isBlocked = true
		}
	}
	return isBlocked, strings.Join(messages, ",")
}

func isFriend(relationships relationships) (bool, string) {
	messages := []string{}
	isFriend := false
	for _, relationship := range relationships {
		if relationship.status == relationshipIsFriend {
			messages = append(messages, relationship.requestor+" is already a friend of  "+relationship.target)
			isFriend = true
		}
	}
	return isFriend, strings.Join(messages, ",")
}

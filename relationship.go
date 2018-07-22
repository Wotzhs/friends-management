package main

import (
	"errors"
	"fmt"
	"regexp"
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

	for _, users := range users {
		if !isEmailValid(users) {
			return errors.New("invalid email being submitted")
		}
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

func getFriendsList(user string) (friends []string, count int, err error) {
	if !isEmailValid(user) {
		err = errors.New("invalid user")
		return
	}

	query := `
		SELECT requestor_relationships.target target FROM relationships requestor_relationships
		LEFT JOIN relationships target_relationships ON requestor_relationships.target = target_relationships.requestor
		WHERE requestor_relationships.requestor=$1 AND target_relationships.target=$1
		AND requestor_relationships.status=$2 AND target_relationships.status = $2
	`

	rows, err := db.Query(query, user, relationshipIsFriend)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to check if user %v has any friends err %v", user, err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		row := relationship{}
		err = rows.Scan(&row.target)
		if err != nil {
			return
		}
		friends = append(friends, row.target)
	}

	count = len(friends)

	if count == 0 {
		err = errors.New("user doesn't have any friends")
		return
	}

	return
}

func getCommonFriendsList(users []string) (friends []string, count int, err error) {
	if len(users) != 2 {
		err = errors.New("incorrect number of friends")
		return
	}
	for _, user := range users {
		if !isEmailValid(user) {
			err = errors.New("invalid user")
			return
		}
	}

	exists, relationships, err := ifExistsRelationship(users)
	if err != nil {
		return
	}

	if exists {
		if isBlocked, message := isBlocked(relationships); isBlocked {
			err = errors.New(message)
			return
		}
	}

	query := `
		/* 
			a = requestors_relationship (user 1 and user 2 relationship)
			b = requestors_target_relationship (user 2 and user 1 relationship)
			c = common_relationship (user 2 and user 3 relationship)
			d = common_requestor_relationship (user 3 and user 1 relationship)

			these requires all related parties to have status of "friend", "blocked" or "subscribed" will not match
			table alias names are intentionally kept short to maintain readability
		*/
		
		SELECT d.target
		FROM 
			relationships a
		INNER JOIN 
			(SELECT requestor, target FROM relationships WHERE requestor = $1) b ON a.target = b.requestor 
			AND a.requestor = b.target
		INNER JOIN 
			relationships c ON c.requestor = b.target
			AND c.status = $3
		INNER JOIN 
			relationships d ON d.requestor = c.target 
			AND d.target = a.requestor 
			AND d.status = $3
		WHERE 
			d.requestor = $2
	`

	rows, err := db.Query(query, users[0], users[1], relationshipIsFriend)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to check if user %v and user %v has any common friends err %v", users[0], users[1], err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		row := relationship{}
		err = rows.Scan(&row.target)
		if err != nil {
			return
		}
		friends = append(friends, row.target)
	}

	count = len(friends)

	if count == 0 {
		err = errors.New("users doesn't have any common friends")
		return
	}

	return
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
			messages = append(messages, relationship.requestor+" is already a friend of "+relationship.target)
			isFriend = true
		}
	}
	return isFriend, strings.Join(messages, ",")
}

func isEmailValid(email string) bool {
	// credit: http://www.golangprograms.com/golang-package-examples/regular-expression-to-validate-email-address.html
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return re.MatchString(email)
}

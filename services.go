package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	relationshipIsFriend     = "friend"
	relationshipIsBlocked    = "blocked"
	relationshipIsSubscribed = "subscribed"
)

func createFriends(users []string) error {
	insertQuery := `
		INSERT INTO relationships (requestor, target, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	user1 := strings.ToLower(users[0])
	user2 := strings.ToLower(users[1])
	if _, err := db.Exec(insertQuery, user1, user2, relationshipIsFriend, now, now); err != nil {
		return err
	}

	if _, err := db.Exec(insertQuery, user2, user1, relationshipIsFriend, now, now); err != nil {
		return err
	}
	return nil
}

func getFriendsList(user string) (friends []string, err error) {
	query := `
		SELECT requestor_relationships.target target FROM relationships requestor_relationships
		LEFT JOIN relationships target_relationships ON requestor_relationships.target = target_relationships.requestor
		WHERE requestor_relationships.requestor=$1 AND target_relationships.target=$1
		AND requestor_relationships.status=$2 AND target_relationships.status = $2
	`

	rows, err := db.Query(query, strings.ToLower(user), relationshipIsFriend)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to check if user %v has any friends err %v", user, err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		row := relationship{}
		err = rows.Scan(&row.Target)
		if err != nil {
			return
		}
		friends = append(friends, row.Target)
	}

	if len(friends) == 0 {
		err = errors.New("user doesn't have any friends")
		return
	}

	return
}

func getCommonFriendsList(users []string) (friends []string, err error) {
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

	rows, err := db.Query(query, strings.ToLower(users[0]), strings.ToLower(users[1]), relationshipIsFriend)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to check if user %v and user %v has any common friends err %v", users[0], users[1], err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		row := relationship{}
		err = rows.Scan(&row.Target)
		if err != nil {
			return
		}
		friends = append(friends, row.Target)
	}

	if len(friends) == 0 {
		err = errors.New("users doesn't have any common friends")
		return
	}

	return
}

func subscribeUpdates(requestor, target string) error {
	subscribeQuery := `
		INSERT INTO relationships (requestor, target, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	if _, err := db.Exec(subscribeQuery, requestor, target, relationshipIsSubscribed, now, now); err != nil {
		return err
	}

	return nil
}

func blockUpdates(requestor, target string) error {
	blockQuery := `
		INSERT INTO relationships (requestor, target, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	if _, err := db.Exec(blockQuery, requestor, target, relationshipIsBlocked, now, now); err != nil {
		return err
	}

	return nil
}

func blockExistingRelationship(requestor, target string) error {
	blockQuery := `
		UPDATE relationships 
		SET status = $1, updated_at = $2
		WHERE requestor = $3 AND target = $4
	`
	now := time.Now()
	if _, err := db.Exec(blockQuery, relationshipIsBlocked, now, requestor, target); err != nil {
		return err
	}

	return nil
}

func getSubscribedList(sender string) (subscribers []string, err error) {
	subscriberQuery := `
		/*
			target_relationship.status may be null because subscription is not set two ways, unlike friendships
			i.e. user A subscribe to user B will not result in user B subscribe to user A
		*/

		SELECT 
			(CASE
				WHEN target_relationship.status IS NOT NULL 
					AND target_relationship.status <> $2 THEN requestor_relationships.requestor
				WHEN requestor_relationships.status = $3 
					AND target_relationship.status IS NULL THEN requestor_relationships.requestor
			END) requestor 
		FROM 
			relationships requestor_relationships
		LEFT JOIN 
			relationships target_relationship ON target_relationship.requestor = requestor_relationships.target
			AND target_relationship.target = requestor_relationships.requestor
		WHERE 
			requestor_relationships.target = $1 
			AND (requestor_relationships.status = $3 OR requestor_relationships.status = $4)
	`

	rows, err := db.Query(subscriberQuery, sender, relationshipIsBlocked, relationshipIsSubscribed, relationshipIsFriend)
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to check if sender %v has any subscribers err %v", sender, err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		row := relationship{}
		err = rows.Scan(&row.Requestor)
		if err != nil {
			return
		}
		subscribers = append(subscribers, row.Requestor)
	}

	return
}

func ifExistsRelationship(users []string) (exists bool, relationships relationships, err error) {
	statusQuery := `
		SELECT requestor, target, status FROM relationships 
		WHERE (requestor=$1 AND target=$2)
		OR (requestor=$2 AND target=$1)
	`

	rows, err := db.Query(statusQuery, strings.ToLower(users[0]), strings.ToLower(users[1]))
	if err != nil {
		err = errors.New(fmt.Sprintf("failed to check if any relationships exists between the users %v", err))
		return
	}

	defer rows.Close()

	for rows.Next() {
		row := relationship{}
		err = rows.Scan(&row.Requestor, &row.Target, &row.Status)
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

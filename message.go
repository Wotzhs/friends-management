package main

import (
	"errors"
	"regexp"
	"strings"
)

type message struct {
	Sender string
	Text   string
}

func (m message) getSubscribers() (user user, err error) {
	if m.Sender == "" {
		err = errors.New("invalid message")
		return
	}

	sender := strings.ToLower(m.Sender)

	// extract all the mentioned users in the text, if any
	emailFilter := regexp.MustCompile(`\S*@\S*`)
	mentionedUsers := emailFilter.FindAllString(m.Text, -1)
	for _, mentionedUser := range mentionedUsers {
		mentionedUser = strings.ToLower(mentionedUser)

		if strings.Contains(mentionedUser, ",") {
			mentionedUser = strings.Replace(mentionedUser, ",", "", -1)
		}

		if isEmailValid(mentionedUser) {
			user.Subscribers = append(user.Subscribers, mentionedUser)
		}
	}

	subscribers, err := getSubscribedList(sender)
	if err != nil && user.Subscribers == nil {
		return
	}
	user.Subscribers = append(user.Subscribers, subscribers...)
	return
}

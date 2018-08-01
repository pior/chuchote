package chat

import (
	"encoding/json"
	"fmt"
)

// CLIENT events

type clientEvent struct {
	Kind    string
	Payload interface{}
}

type clientEventMessage struct {
	Body string
}

func parseClientEvent(data []byte) (interface{}, error) {
	event := &clientEvent{}
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	if event.Kind == "message" {
		event = &clientEvent{Payload: &clientEventMessage{}}
	} else {
		return nil, fmt.Errorf("invalid client event kind %s", event.Kind)
	}

	err = json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}

	return event.Payload, nil
}

// SERVER events

type serverEvent struct {
	Kind    string
	Payload interface{}
}

type serverEventMembers struct {
	Members []string
}

type serverEventMessage struct {
	From clientID
	Body string
}

func newEventMessage(from clientID, body string) *serverEvent {
	return &serverEvent{
		Kind:    "message",
		Payload: &serverEventMessage{From: from, Body: body},
	}
}

func newEventMembers(members []string) *serverEvent {
	return &serverEvent{
		Kind:    "members",
		Payload: &serverEventMembers{Members: members},
	}
}

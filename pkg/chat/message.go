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

type serverEventClientState struct {
	Name string
}

type serverEventRoomState struct {
	Name    string
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

func newEventClientState(name string) *serverEvent {
	return &serverEvent{
		Kind:    "clientState",
		Payload: &serverEventClientState{Name: name},
	}
}

func newEventRoomState(name string, members []string) *serverEvent {
	return &serverEvent{
		Kind:    "roomState",
		Payload: &serverEventRoomState{Name: name, Members: members},
	}
}

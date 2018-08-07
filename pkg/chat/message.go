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
		return nil, fmt.Errorf("invalid client event: %+v", event)
	}

	err = json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}

	return event.Payload, nil
}

// SERVER events

type clientInfo struct {
	Name string
	id   clientID
}

type serverEvent struct {
	Kind    string
	From    *clientInfo
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
	Body string
}

func newEventMessage(fromID clientID, fromName string, body string) *serverEvent {
	return &serverEvent{
		Kind:    "message",
		From:    &clientInfo{id: fromID, Name: fromName},
		Payload: &serverEventMessage{Body: body},
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

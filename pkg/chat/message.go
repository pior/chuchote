package chat

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

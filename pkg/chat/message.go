package chat

type eventInfo struct {
	Kind        string
	Description string
}

type eventMessage struct {
	From clientID
	Body string
}

type event struct {
	Message *eventMessage
	Info    *eventInfo
}

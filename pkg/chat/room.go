package chat

import (
	"encoding/json"
	"fmt"

	"github.com/dchest/uniuri"
	"github.com/gorilla/websocket"
)

type roomID string

func RandomRoomID() roomID {
	return roomID(uniuri.New())
}

type room struct {
	id      roomID
	core    *core
	channel chan *serverEvent
	clients map[clientID]*client
}

func newRoom(core *core, id roomID) *room {
	r := &room{
		id:      id,
		core:    core,
		channel: make(chan *serverEvent),
		clients: make(map[clientID]*client),
	}
	go r.broadcaster()
	return r
}

func (r *room) broadcaster() {
	for event := range r.channel {
		payload, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("Error serializing: %+v : %s", event, err)
		}

		for _, client := range r.clients {
			if event.From != nil && event.From.id == client.id {
				continue
			}
			client.channel <- payload
		}
	}
}

func (r *room) broadcast(event *serverEvent) {
	r.channel <- event
}

func (r *room) createClient(conn *websocket.Conn) {
	c := newClient(conn, r)
	r.clients[c.id] = c

	r.pushMemberList()
}

func (r *room) pushMemberList() {
	var members []string
	for _, client := range r.clients {
		members = append(members, client.name)
	}

	r.broadcast(newEventRoomState(string(r.id), members))
}

func (r *room) deleteClient(cl *client) {
	delete(r.clients, cl.id)

	if len(r.clients) == 0 {
		r.core.deleteRoom(r)
		return
	}

	r.pushMemberList()
}

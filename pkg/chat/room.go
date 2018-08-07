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
	channel chan []byte
	clients map[clientID]*client
}

func newRoom(core *core, id roomID) *room {
	r := &room{
		id:      id,
		core:    core,
		channel: make(chan []byte),
		clients: make(map[clientID]*client),
	}
	go r.broadcaster()
	return r
}

func (r *room) broadcaster() {
	for msg := range r.channel {
		for _, client := range r.clients {
			client.channel <- msg
		}
	}
}

func (r *room) broadcast(event *serverEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Error sending: %s", err)
	}
	r.channel <- payload
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

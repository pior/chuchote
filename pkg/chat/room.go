package chat

import (
	"encoding/json"
	"fmt"

	"github.com/cskr/pubsub"
	"github.com/gorilla/websocket"
	"github.com/teris-io/shortid"
)

type roomID string

func (roomID) NewRandom() roomID {
	return roomID(shortid.MustGenerate())
}

type room struct {
	id      roomID
	core    *core
	pubsub  *pubsub.PubSub
	clients map[clientID]*client
}

func newRoom(core *core, id roomID) *room {
	return &room{
		id:      id,
		core:    core,
		pubsub:  pubsub.New(10),
		clients: make(map[clientID]*client),
	}
}

func (r *room) sendMessage(message string) {
	r.pubsub.Pub(message, "msg")
}

func (r *room) send(event *event) {
	payload, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Error sending: %s", err)
	}
	r.pubsub.Pub(string(payload), "msg")
}

func (r *room) sendInfo(message string) {
	r.pubsub.Pub(fmt.Sprintf("info: %s", message), "msg")
}

func (r *room) getMessageChannel() chan interface{} {
	return r.pubsub.Sub("msg")
}

func (r *room) closeMessageChannel(ch chan interface{}) {
	r.pubsub.Unsub(ch)
}

func (r *room) createClient(conn *websocket.Conn) *client {
	c := newClient(conn, r)
	r.clients[c.id] = c
	r.send()
	r.sendInfo(fmt.Sprintf("new client list: %+v", r.clients))
	return c
}

func (r *room) deleteClient(cl *client) {
	delete(r.clients, cl.id)

	if len(r.clients) == 0 {
		r.close()
		r.core.deleteRoom(r)
		return
	}

	r.sendInfo(fmt.Sprintf("new client list: %+v", r.clients))
}

func (r *room) close() {
	r.pubsub.Shutdown()
}

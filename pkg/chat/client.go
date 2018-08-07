package chat

import (
	"fmt"

	"github.com/dchest/uniuri"
	"github.com/dustinkirkland/golang-petname"
	"github.com/gorilla/websocket"
)

type clientID string

type client struct {
	id      clientID
	name    string
	conn    *websocket.Conn
	room    *room
	channel chan []byte
}

func newClient(conn *websocket.Conn, room *room) *client {
	c := &client{
		id:      clientID(uniuri.New()),
		name:    petname.Generate(1, ""),
		conn:    conn,
		room:    room,
		channel: make(chan []byte),
	}

	go c.reader()
	go c.writer()

	return c
}

func (c *client) close() {
	fmt.Println("Deleting client: ", c.id, c.name)
	c.room.deleteClient(c)

	close(c.channel)

	fmt.Println("Closing connection: ", c.conn.RemoteAddr())
	c.conn.Close()
}

func (c *client) reader() {
	defer c.close()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("%s: Received from %s: \"%s\"\n", c.id, c.conn.RemoteAddr(), message)

		c.processMessage(message)
	}
	fmt.Printf("%s: stopping reader\n", c.id)
}

func (c *client) processMessage(message []byte) {
	event, err := parseClientEvent(message)
	if err != nil {
		fmt.Printf("error parsing client event: %s", err)
	}

	switch e := event.(type) {
	case *clientEventMessage:
		c.room.broadcast(newEventMessage(c.id, c.name, e.Body))
	default:
		fmt.Printf("unknown event type: %T : %+v", e, e)
	}
}

// writer is the only place that writes to the connection
func (c *client) writer() {
	for message := range c.channel {
		fmt.Printf("%s: writing to %s: \"%s\"\n", c.id, c.conn.RemoteAddr(), message)
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Printf("%s: error writing to %s: %s\n", c.id, c.conn.RemoteAddr(), err)
			break
		}
	}

	fmt.Printf("%s: writer stopped\n", c.id)
}

func (c *client) sendMessage(message []byte) {
	c.channel <- message
}

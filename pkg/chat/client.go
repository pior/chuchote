package chat

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/teris-io/shortid"
)

type clientID string

type client struct {
	id   clientID
	conn *websocket.Conn
	room *room
}

func newClient(conn *websocket.Conn, room *room) *client {
	return &client{
		id:   clientID(shortid.MustGenerate()),
		conn: conn,
		room: room,
	}
}

func (c *client) reader() {
	defer func() {
		c.conn.Close()
		c.room.deleteClient(c)
	}()

	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("%s: Received from %s: \"%s\"\n", c.id, c.conn.RemoteAddr(), p)

		c.room.send(&event{
			Message: &eventMessage{
				From: c.id,
				Body: string(p),
			},
		})
	}
}

func (c *client) writer() {
	ch := c.room.getMessageChannel()
	defer c.room.closeMessageChannel(ch)

	for msg := range ch {
		fmt.Printf("%s: Sending to %s: \"%s\"\n", c.id, c.conn.RemoteAddr(), msg)
		err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg.(string)))
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/cskr/pubsub"

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
		c.room.sendMessage(string(p))
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

type room struct {
	id      string
	core    *core
	pubsub  *pubsub.PubSub
	clients map[clientID]*client
}

func newRoom(core *core, id string) *room {
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

type core struct {
	upgrader *websocket.Upgrader
	rooms    map[string]*room
}

func newCore() *core {
	return &core{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		rooms: make(map[string]*room),
	}
}

func (co *core) getOrCreateRoom(id string) *room {
	if _, present := co.rooms[id]; !present {
		co.rooms[id] = newRoom(co, id)
	}
	return co.rooms[id]
}

func (co *core) connect(conn *websocket.Conn, roomID string) {
	room := co.getOrCreateRoom(roomID)
	cl := room.createClient(conn)
	go cl.reader()
	go cl.writer()
}

func (co *core) deleteRoom(room *room) {
	delete(co.rooms, room.id)
	fmt.Printf("Deleted a room: %+v\n", room)
	fmt.Printf("New list of rooms: %+v\n", co.rooms)
}

func (co *core) serve(c echo.Context) error {
	roomID := c.Param("id")
	if roomID == "" {
		return c.String(http.StatusBadRequest, "invalid room id")
	}

	conn, err := co.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "failed to upgrade to websocket connection")
	}

	co.connect(conn, roomID)
	return nil
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		roomID := shortid.MustGenerate()
		return c.Redirect(http.StatusTemporaryRedirect, "/r/"+roomID)
	})

	e.GET("/r/:id", func(c echo.Context) error {
		return c.File("public/index.html")
	})

	c := newCore()
	e.GET("/r/:id/socket", c.serve)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.Fatal(e.Start(":8000"))
}

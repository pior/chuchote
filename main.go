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

func generateRandomRoomID() string {
	return shortid.MustGenerate()
}

type emitter func(string)

type closer func()

type client struct {
	conn       *websocket.Conn
	emit       emitter
	receiverCh chan interface{}
	close      closer
}

func (c *client) reader() {
	defer c.close()

	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("Received from %s: %s\n", c.conn.RemoteAddr(), p)
		c.emit(string(p))
	}
}

func (c *client) writer() {
	for msg := range c.receiverCh {
		fmt.Printf("Sending to %s: %s\n", c.conn.RemoteAddr(), msg)
		err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg.(string)))
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

type core struct {
	pubsub   *pubsub.PubSub
	upgrader *websocket.Upgrader
}

func newCore() *core {
	return &core{
		pubsub: pubsub.New(10),
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (co *core) startClient(conn *websocket.Conn, room string) {
	inboundCh := co.pubsub.Sub(room)

	cl := &client{
		conn: conn,
		close: func() {
			conn.Close()
			co.pubsub.Unsub(inboundCh)
		},
		emit: func(s string) {
			co.pubsub.Pub(s, room)
		},
		receiverCh: inboundCh,
	}

	go cl.reader()
	go cl.writer()
}

func (co *core) Serve(c echo.Context) error {
	roomID := c.Param("id")
	if roomID == "" {
		return c.String(http.StatusBadRequest, "invalid room id")
	}

	conn, err := co.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "failed to upgrade to websocket connection")
	}

	co.startClient(conn, roomID)
	return nil
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/r/%s", generateRandomRoomID()))
	})

	e.GET("/r/:id", func(c echo.Context) error {
		return c.File("public/index.html")
	})

	c := newCore()
	e.GET("/r/:id/socket", c.Serve)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.Fatal(e.Start(":8000"))
}

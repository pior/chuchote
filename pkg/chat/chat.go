package chat

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/labstack/echo"
)

type core struct {
	upgrader *websocket.Upgrader
	rooms    map[roomID]*room
}

func NewCore() *core {
	return &core{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		rooms: make(map[roomID]*room),
	}
}

func (co *core) getOrCreateRoom(id string) *room {
	rID := roomID(id)
	if _, present := co.rooms[rID]; !present {
		co.rooms[rID] = newRoom(co, rID)
	}
	return co.rooms[rID]
}

func (co *core) connect(conn *websocket.Conn, roomID string) {
	room := co.getOrCreateRoom(roomID)
	room.createClient(conn)
}

func (co *core) deleteRoom(room *room) {
	delete(co.rooms, room.id)
	fmt.Printf("Deleted a room: %s\n", room.id)
	fmt.Printf("New list of rooms: %+v\n", co.rooms)
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

	co.connect(conn, roomID)
	return nil
}

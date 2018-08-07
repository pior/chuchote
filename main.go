package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/pior/chuchote/pkg/chat"
)

func getBindString() string {
	portVar := os.Getenv("PORT")
	if portVar == "" {
		portVar = "8000"
	}
	port, err := strconv.Atoi(portVar)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(":%d", port)
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		roomID := chat.RandomRoomID()
		return c.Redirect(http.StatusTemporaryRedirect, "/r/"+string(roomID))
	})

	e.GET("/r/:id", func(c echo.Context) error {
		return c.File("public/index.html")
	})

	c := chat.NewCore()
	e.GET("/r/:id/socket", c.Serve)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Logger.Fatal(e.Start(getBindString()))
}

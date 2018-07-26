package main

import (
	"net/http"

	"github.com/teris-io/shortid"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/pior/chuchote/pkg/chat"
)

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		roomID := shortid.MustGenerate()
		return c.Redirect(http.StatusTemporaryRedirect, "/r/"+roomID)
	})

	e.GET("/r/:id", func(c echo.Context) error {
		return c.File("public/index.html")
	})

	c := chat.NewCore()
	e.GET("/r/:id/socket", c.Serve)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.Fatal(e.Start(":8000"))
}

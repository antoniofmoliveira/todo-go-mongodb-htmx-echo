package main

import (
	"example.me/todo/services"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	dbService := services.NewDB()
	defer dbService.Close()
	todoService := &services.TodoService{Collection: dbService.Collection}
	handlerService := &services.HandlerService{TodoService: *todoService}

	e := echo.New()
	e.Renderer = services.NewTemplates()
	e.Use(middleware.Logger())
	e.Static("/images", "static/images")
	e.Static("/css", "static/css")
	e.Static("/js", "static/js")

	e.GET("/", func(c echo.Context) error {
		return handlerService.AllTodosHandler(c)
	})

	e.POST("/", func(c echo.Context) error {
		return handlerService.CreateTodoHandler(c)
	})

	e.DELETE("/:id", func(c echo.Context) error {
		return handlerService.DeleteTodoHandler(c)
	})

	e.GET("/toggle/:id", func(c echo.Context) error {
		return handlerService.ToggleTodoHandler(c)
	})

	e.Logger.Fatal(e.Start(":42069"))
}

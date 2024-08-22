package services

import (
	"net/http"

	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HandlerService struct {
	TodoService TodoService
}

func (h *HandlerService) AllTodosHandler(c echo.Context) error {
	todos, err := h.TodoService.AllTodos()
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.Render(http.StatusOK, "index", todos)
}

func (h *HandlerService) CreateTodoHandler(c echo.Context) error {
	todo := Todo{Title: c.FormValue("title")}
	insertedId, err := h.TodoService.AddTodo(todo)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	todo, err = h.TodoService.GetTodo(insertedId)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	c.Render(http.StatusOK, "input", interface{}(nil))
	return c.Render(http.StatusOK, "todo", todo)
}

func (h *HandlerService) DeleteTodoHandler(c echo.Context) error {
	oid, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	_, err = h.TodoService.DeleteTodo(oid)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *HandlerService) ToggleTodoHandler(c echo.Context) error {

	oid, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	todo, err := h.TodoService.GetTodo(oid)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	todo.Done = !todo.Done
	_, err = h.TodoService.UpdateTodo(todo)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.Render(http.StatusOK, "todo", todo)
}

package main

import (
	"context"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
The `Templates` struct in your code is a wrapper around the
`template.Template` type. It provides a way to store and manage
a collection of templates.

Here's a brief explanation of each method in the `Templates`
struct:

- `Render`: This method is used to render a specific template
with the given data. It takes an `io.Writer` to write the
rendered template to, a `name` string to specify the template
to render, `data` of any type to provide data to the template,
and an `echo.Context` to provide additional context to the
template. It returns an `error` if there was a problem
rendering the template.

- `NewTemplates`: This method creates a new instance of the
`Templates` struct. It takes no parameters and returns a
pointer to a `Templates` instance.

The `Templates` struct is a simple wrapper around the
`template.Template` type, providing a convenient way to manage
and render templates in your application.
*/
type Templates struct {
	templates *template.Template
}

// Render renders a template with the given data.
//
// The function takes an io.Writer, a template name, an
// interface{} data, and an echo.Context.
// It returns an error.
func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// NewTemplates returns a new instance of Templates.
//
// No parameters.
// Returns a pointer to Templates.
func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
}

type Todo struct {
	ID    primitive.ObjectID `bson:"_id"`
	Title string
	Done  bool
}

type TodoService struct {
	collection *mongo.Collection
}

func (s *TodoService) AllTodos() ([]Todo, error) {
	cursor, err := s.collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	var results []Todo
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *TodoService) GetTodo(id interface{}) (Todo, error) {
	var result Todo
	err := s.collection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&result)
	if err != nil {
		return Todo{}, err
	}
	return result, nil
}

func (s *TodoService) AddTodo(t Todo) (interface{}, error) {
	document := bson.D{{Key: "title", Value: t.Title}, {Key: "done", Value: false}}
	result, err := s.collection.InsertOne(context.TODO(), document)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (s *TodoService) UpdateTodo(t Todo) (interface{}, error) {
	filter := bson.D{{Key: "_id", Value: t.ID}}
	// update := bson.D{{Key: "$set", Value: bson.D{{Key: "title", Value: t.title}, {Key: "done", Value: t.done}}}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "done", Value: t.Done}}}}
	result, err := s.collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return nil, err
	}
	return result.UpsertedID, nil
}

func (s *TodoService) DeleteTodo(id interface{}) (interface{}, error) {
	filter := bson.D{{Key: "_id", Value: id}}
	result, err := s.collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type HandlerService struct {
	todoService TodoService
}

func (h *HandlerService) AllTodosHandler(c echo.Context) error {
	todos, err := h.todoService.AllTodos()
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.Render(http.StatusOK, "index", todos)
}

func (h *HandlerService) CreateTodoHandler(c echo.Context) error {
	todo := Todo{Title: c.FormValue("title")}
	insertedId, err := h.todoService.AddTodo(todo)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	todo, err = h.todoService.GetTodo(insertedId)
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
	_, err = h.todoService.DeleteTodo(oid)
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
	todo, err := h.todoService.GetTodo(oid)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	todo.Done = !todo.Done
	_, err = h.todoService.UpdateTodo(todo)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.Render(http.StatusOK, "todo", todo)
}

// main is the entry point of the Go program.
//
// It initializes the environment, sets up the MongoDB connection, and starts the Echo server.
// No parameters.
// No return values.
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " +
			"www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	collection := client.Database("todo_go").Collection("todo_go")
	todoService := &TodoService{collection: collection}
	handlerService := &HandlerService{todoService: *todoService}

	e := echo.New()
	e.Renderer = NewTemplates()
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

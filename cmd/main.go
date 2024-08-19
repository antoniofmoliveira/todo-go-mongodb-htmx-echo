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

/*
This is not a class definition, but a struct definition in Go.
Here's a succinct explanation:

**Todo Struct**

* The Todo struct represents a Todo item with three fields:
  - `ID`: a unique identifier of type `primitive.ObjectID`,
    which is used in MongoDB.
  - `Title`: a string representing the title of the Todo item.
  - `Done`: a boolean indicating whether the Todo item is
    done or not.

Note that this struct does not have any methods, only fields.
*/
type Todo struct {
	ID    primitive.ObjectID `bson:"_id"`
	Title string
	Done  bool
}

// AllTodos retrieves all Todo items from the MongoDB collection.
//
// It takes a pointer to a mongo.Collection as a parameter.
// Returns a slice of Todo items and an error.
func AllTodos(collection *mongo.Collection) ([]Todo, error) {
	cursor, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	var results []Todo
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

// GetTodo retrieves a Todo item from the MongoDB collection by its ID.
//
// It takes a pointer to a mongo.Collection and an ID as parameters.
// Returns a Todo item and an error.
func GetTodo(collection *mongo.Collection, id interface{}) (Todo, error) {
	var result Todo
	err := collection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&result)
	if err != nil {
		return Todo{}, err
	}
	return result, nil
}

// AddTodo adds a new Todo item to the MongoDB collection.
//
// It takes a pointer to a mongo.Collection and a Todo item as parameters.
// Returns the ID of the newly inserted Todo item and an error.
func AddTodo(collection *mongo.Collection, t Todo) (interface{}, error) {
	document := bson.D{{Key: "title", Value: t.Title}, {Key: "done", Value: false}}
	result, err := collection.InsertOne(context.TODO(), document)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

// UpdateTodo updates a Todo item in the MongoDB collection.
//
// It takes a pointer to a mongo.Collection and a Todo item as parameters.
// Returns the ID of the updated Todo item and an error.
func UpdateTodo(collection *mongo.Collection, t Todo) (interface{}, error) {
	filter := bson.D{{Key: "_id", Value: t.ID}}
	// update := bson.D{{Key: "$set", Value: bson.D{{Key: "title", Value: t.title}, {Key: "done", Value: t.done}}}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "done", Value: t.Done}}}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return nil, err
	}
	return result.UpsertedID, nil
}

// DeleteTodo deletes a Todo item from the MongoDB collection.
//
// It takes a pointer to a mongo.Collection and an ID as parameters.
// Returns the result of the deletion operation and an error.
func DeleteTodo(collection *mongo.Collection, id interface{}) (interface{}, error) {
	filter := bson.D{{Key: "_id", Value: id}}
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AllTodosHandler handles the retrieval of all Todo items from the MongoDB collection.
//
// It takes an echo.Context and a pointer to a mongo.Collection as parameters.
// Returns an error.
func AllTodosHandler(c echo.Context, collection *mongo.Collection) error {
	todos, err := AllTodos(collection)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.Render(http.StatusOK, "index", todos)
}

// CreateTodoHandler handles the creation of a new Todo item.
//
// It takes an echo.Context and a pointer to a mongo.Collection as parameters.
// Returns an error.
func CreateTodoHandler(c echo.Context, collection *mongo.Collection) error {
	todo := Todo{Title: c.FormValue("title")}
	insertedId, err := AddTodo(collection, todo)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	todo, err = GetTodo(collection, insertedId)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	c.Render(http.StatusOK, "input", interface{}(nil))
	return c.Render(http.StatusOK, "todo", todo)
}

// DeleteTodoHandler handles the deletion of a Todo item from the MongoDB collection.
//
// It takes an echo.Context and a pointer to a mongo.Collection as parameters.
// Returns an error.
func DeleteTodoHandler(c echo.Context, collection *mongo.Collection) error {
	oid, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	_, err = DeleteTodo(collection, oid)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

// ToggleTodoHandler updates the status of a Todo item in the MongoDB collection.
//
// It takes an echo.Context and a pointer to a mongo.Collection as parameters.
// Returns an error.
func ToggleTodoHandler(c echo.Context, collection *mongo.Collection) error {

	oid, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	todo, err := GetTodo(collection, oid)
	if err != nil {
		c.Render(http.StatusOK, "messages", err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}
	todo.Done = !todo.Done
	_, err = UpdateTodo(collection, todo)
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

	e := echo.New()
	e.Renderer = NewTemplates()
	e.Use(middleware.Logger())
	e.Static("/images", "static/images")
	e.Static("/css", "static/css")
	e.Static("/js", "static/js")

	e.GET("/", func(c echo.Context) error {
		return AllTodosHandler(c, collection)
	})

	e.POST("/", func(c echo.Context) error {
		return CreateTodoHandler(c, collection)
	})

	e.DELETE("/:id", func(c echo.Context) error {
		return DeleteTodoHandler(c, collection)
	})

	e.GET("/toggle/:id", func(c echo.Context) error {
		return ToggleTodoHandler(c, collection)
	})

	e.Logger.Fatal(e.Start(":42069"))
}

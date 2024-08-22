package services

import (
	"html/template"
	"io"

	"github.com/labstack/echo"
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

package view

import (
	"html/template"
	"io"
	"log"
	"path"
	"strings"
)

// Development enables recompilation of templates for easier development.
var Development = false

var funcs = template.FuncMap{
	"markdown":    markdown,
	"highlight":   highlight,
	"joinURL":     path.Join,
	"baseURL":     path.Base,
	"breadcrumbs": breadcrumbs,
}

var templates = template.Must(template.New("index").Funcs(funcs).ParseGlob("web/*.html"))

// Render writes a template with the given name and data to the writer.
func Render(w io.Writer, name string, data interface{}) {
	if Development {
		templates = template.Must(template.New("index.html").Funcs(funcs).ParseGlob("web/*.html"))
	}

	var b strings.Builder
	if err := templates.ExecuteTemplate(&b, name, data); err != nil {
		log.Println(err)
	}

	if err := templates.Execute(w, template.HTML(b.String())); err != nil {
		log.Println(err)
	}
}

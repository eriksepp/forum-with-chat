package view

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"forum/logger"
)

// We might want to make a configuration file for these hardcoded values.
const (
	TEMPLATES_PATH = "./webui/templates/"
	STATIC_PATH    = "./webui/static/"
)

var (
	Files = []string{
		TEMPLATES_PATH + "base.html",
		TEMPLATES_PATH + "login.html",
		TEMPLATES_PATH + "register.html",
		TEMPLATES_PATH + "navbar.html",
		TEMPLATES_PATH + "createpost.html",
		TEMPLATES_PATH + "onlineusers.html",
		TEMPLATES_PATH + "postslist.html",
		TEMPLATES_PATH + "chat.html",
		TEMPLATES_PATH + "fullpost.html",
	} // Files that are always served.
)

type View struct {
	Template *template.Template
}

// Creates a new template view
func New(path string) (*View, error) {
	var view View
	var err error

	files := append(make([]string, 0), TEMPLATES_PATH+path)
	files = append(files, Files...)

	// Load the templates
	view.Template, err = template.ParseFiles(files...)
	if err != nil {
		return nil, errors.New(logger.GetCurrentFuncName() + ": " + err.Error())
	}

	return &view, nil
}

// Executes the template and sends it to the client
func (v *View) Execute(w http.ResponseWriter, vars map[string]any) error {

	// Execute the template with our custom data (vars)
	return v.Template.Execute(w, vars)
}

/*
executes a template for the given error (statusCode)
*/
func ExecuteError(w http.ResponseWriter, r *http.Request, statusCode int) error {
	var pageName string
	switch statusCode {
	case http.StatusNotFound:
		pageName = "error404.html"
	case http.StatusForbidden:
		pageName = "forbidden.tmpl"
	default:
		pageName = "error404.html"
	}

	view, err := New(pageName)
	if err != nil {
		http.NotFound(w, r)
		return fmt.Errorf("can't parse %s template: %w", pageName, err)
	}

	if err = view.Execute(w, nil); err != nil {
		http.NotFound(w, r)
		return fmt.Errorf("can't execute  %s template: %w", pageName, err)
	}
	return nil
}

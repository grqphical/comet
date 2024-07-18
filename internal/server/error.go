package server

import (
	"comet/internal"
	"html/template"
	"net/http"
)

var ErrorTemplate *template.Template

func init() {
	ErrorTemplate, _ = template.New("error.html").ParseFiles("templates/error.html")
}
func sendError(code int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")

	data := struct {
		Code    int
		Message string
		Version string
	}{
		Code:    code,
		Message: http.StatusText(code),
		Version: internal.Version,
	}

	_ = ErrorTemplate.Execute(w, data)
}

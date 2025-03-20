package handlers

import (
	"fmt"
	"html/template"
	"net/http"
)

type ErrorPage struct {
	StatusCode string
	StatusText string
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, statusCode int, statusText string) {
	w.WriteHeader(statusCode)

	data := ErrorPage{
		StatusCode: fmt.Sprint(statusCode),
		StatusText: statusText,
	}

	ts, err := template.ParseFiles("web/templates/wentwrong.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ts.Execute(w, data)
	if err != nil {
		http.Error(w, "Error when executing", http.StatusInternalServerError)
		return
	}
}

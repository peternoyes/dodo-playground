package main

import (
	"encoding/json"
	"fmt"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func BuildErrorResponse(w http.ResponseWriter, statusCode int, message interface{}) {
	var m string
	switch t := message.(type) {
	case string:
		m = t
	case error:
		m = t.Error()
	case fmt.Stringer:
		m = t.String()
	default:
		m = "Unknown Error"
	}

	fmt.Println(m)

	response := ErrorResponse{m}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func loadTemplates() (*template.Template, error) {
	t := template.New("").Funcs(template.FuncMap{})
	t, err := vfstemplate.ParseGlob(assets, t, "/assets/*.tmpl")
	return t, err
}

func Main(w http.ResponseWriter, req *http.Request) {
	t, err := loadTemplates()
	if err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data = struct {
		Animals string
	}{
		Animals: "gophers",
	}

	err = t.ExecuteTemplate(w, "index.html.tmpl", data)
	if err != nil {
		log.Println("t.Execute:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Build(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		BuildErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if err = r.Body.Close(); err != nil {
		BuildErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	output, err := Compile(body)

	if err != nil {
		BuildErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	res := struct {
		Binary []byte `json:"binary"`
	}{
		output,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}
}

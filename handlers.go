package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
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
	template := "index.html.tmpl"
	var obj interface{} = nil
	if ok, user := authenticated(req); ok {
		template = "index.loggedin.html.tmpl" // If authenticated server the full application, not just playground
		obj = user
	}

	t, err := loadTemplates()
	if err != nil {
		log.Println("loadTemplates:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, template, obj)
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

	source := string(body)
	crc := GenerateCRC(source)

	err = nil
	var output []byte

	l := r.Header.Get("X-Language")
	if l == "" {
		l = "c"
	}

	b, _ := GetBinary(crc)
	if b != nil && b.Language == l && b.Version == "1.0.0" {
		if b.Results == "Success" {
			output = b.Fram
			err = nil
		} else {
			output = nil
			err = errors.New(b.Results)
		}
	} else {
		output, err = Compile(body, l)

		results := ""
		if err != nil {
			results = err.Error()
		} else {
			results = "Success"
		}

		b = &Binary{}
		b.New(crc, source, l, output, results, "1.0.0")

		errStore := StoreBinary(b)

		if errStore != nil {
			BuildErrorResponse(w, http.StatusInternalServerError, errStore)
			return
		}
	}

	if err != nil {
		BuildErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	res := struct {
		Binary []byte `json:"binary"`
		Id     string `json:"id"`
	}{
		output,
		crc,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}
}

func Code(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	crc := vars["id"]

	b, err := GetBinary(crc)
	if err != nil {
		BuildErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	code := ""
	if b != nil {
		code = b.Source
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Language", b.Language)
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, code)
}

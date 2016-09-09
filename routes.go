package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)
		router.Methods(route.Method, "OPTIONS").Path(route.Pattern).Name(route.Name).Handler(handler)
	}

	return router
}

var routes = Routes{
	Route{
		"Main",
		"GET",
		"/",
		Main,
	},
	Route{
		"Build",
		"POST",
		"/build",
		Build,
	},
}

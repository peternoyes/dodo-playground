package main

import (
	"log"
	"net/http"
	"time"
)

func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, X-Language, X-Version")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization, Content-Type, Accept, X-Language, X-Version")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Expires", "-1")

		log.Printf("%s\t%s\t%s\t%s", r.Method, r.RequestURI, name, time.Since(start))

		if r.Method != "OPTIONS" {
			inner.ServeHTTP(w, r)
		}
	})
}

package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func ProjectsList(w http.ResponseWriter, r *http.Request) {
	if ok, user := authenticated(r); ok {
		projects, err := GetProjects(user.Email)
		if err != nil {
			BuildErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		res := make([]*Project, len(projects))
		for i, s := range projects {
			res[i] = &Project{Title: s}
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			panic(err)
		}

	} else {
		BuildErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
	}
}

func ProjectGet(w http.ResponseWriter, r *http.Request) {
	if ok, user := authenticated(r); ok {
		vars := mux.Vars(r)
		title := vars["title"]

		project, err := GetProject(user.Email, title)
		if err != nil {
			BuildErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(project); err != nil {
			panic(err)
		}
	} else {
		BuildErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
	}
}

func ProjectUpdate(w http.ResponseWriter, r *http.Request) {
	if ok, user := authenticated(r); ok {
		vars := mux.Vars(r)
		title := vars["title"]

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

		l := r.Header.Get("X-Language")
		if l == "" {
			l = "c"
		}

		v := r.Header.Get("X-Version")
		if v == "" {
			v = DefaultVersion()
		}

		project := &Project{}
		project.Title = title
		project.Source = source
		project.Language = l
		project.Version = v

		err = StoreProject(user.Email, project)
		if err != nil {
			BuildErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	} else {
		BuildErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
	}
}

func ProjectAdd(w http.ResponseWriter, r *http.Request) {
	if ok, user := authenticated(r); ok {
		vars := mux.Vars(r)
		title := vars["title"]

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

		l := r.Header.Get("X-Language")
		if l == "" {
			l = "c"
		}

		v := r.Header.Get("X-Version")
		if v == "" {
			v = DefaultVersion()
		}

		project := &Project{}
		project.Title = title
		project.Source = source
		project.Language = l
		project.Version = v

		err = CreateProject(user.Email, project)
		if err != nil {
			BuildErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	} else {
		BuildErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
	}
}

func ProjectDelete(w http.ResponseWriter, r *http.Request) {
	if ok, user := authenticated(r); ok {
		vars := mux.Vars(r)
		title := vars["title"]

		err := DeleteProject(user.Email, title)

		if err != nil {
			BuildErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	} else {
		BuildErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
	}
}

package main

import (
	"encoding/json"
	"errors"
	//"fmt"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"io/ioutil"
	"net/http"
	"strings"
)

var activeProject string = ""

func IsProjects() bool {
	return jQuery("#projectslist").Get("length").Int() > 0
}

func projectsLogic() {
	// Test if has the projects section
	if IsProjects() {
		click := js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
			go func() {
				setActiveProject(this)
			}()
			return nil
		})
		jQuery("#projectslist").On(jquery.CLICK, "a", click)

		jQuery("#newButton").On(jquery.CLICK, func() {
			go func() {
				newProject()
			}()
		})

		jQuery("#saveButton").On(jquery.CLICK, func() {
			go func() {
				saveProjectCode()
			}()
		})

		jQuery("#deleteButton").AddClass("disabled")
		jQuery("#deleteButton").SetProp("disabled", true)

		jQuery("#deleteButton").On(jquery.CLICK, func() {
			go func() {
				deleteProject()
			}()
		})

		loadProjectList()
	}
}

func setActiveProject(obj *js.Object) {
	jQuery("#projectslist > a").RemoveClass("active")
	if obj != nil && obj != js.Undefined {
		fmt.Println(obj)
		jQuery(obj).AddClass("active")
		activeProject = jQuery(obj).Text()

		jQuery("#deleteButton").RemoveClass("disabled")
		jQuery("#deleteButton").SetProp("disabled", false)
	} else {
		activeProject = ""
		jQuery("#deleteButton").AddClass("disabled")
		jQuery("#deleteButton").SetProp("disabled", true)
	}
	loadProjectCode(activeProject)
}

func loadProjectList() {
	jQuery("#projectslist").Children("").Remove()

	projects, err := getProjects()
	if err == nil {
		activeID := ""
		for i, p := range projects {
			if i == 0 {
				activeID = p
			}

			jQuery("#projectslist").Append("<a id='project" + p + "' href='#' class='list-group-item'>" + p + "</a>")
		}

		var active *js.Object = nil
		if activeID != "" {
			active = jQuery("#project" + activeID).Get()
		}

		setActiveProject(active)
	}
}

func newProject() bool {
	title := jQuery("#newBox").Val()
	if title == "" {
		setStatus("Project name cannot be blank", "bg-danger")
		return false
	}

	// Sample Application
	source := `#include <stdio.h>
#include <stdlib.h>
#include "api.h"

int main() {
	// Initialize the API
	api_init();

	// Clear the graphics in video memory
	CLEAR();

	for (;;) {
		// Game Logic
		
		SET_CURSOR(0, 0);
		DRAW_STRING("Hello, World!");

		// Push video memory to the OLED
		DISPLAY();

		// Wait for next interrupt
		WAIT();
	}

	return 0;
}`

	reader := strings.NewReader(source)

	response, err := http.Post("/projects/"+title, "application/text", reader)

	defer response.Body.Close()
	_, err = ioutil.ReadAll(response.Body)
	if err == nil {
		if response.StatusCode == http.StatusOK {
			setStatus("Succesfully created "+title, "bg-success")
			jQuery("#projectslist").Append("<a id='project" + title + "' href='#' class='list-group-item'>" + title + "</a>")
			jQuery("#newBox").SetVal("")

			setActiveProject(jQuery("#project" + title).Get())
		} else {
			setStatus("Failed to create "+title, "bg-danger")
			return false
		}
	} else {
		setStatus("Failed to create "+title, "bg-danger")
		return false
	}

	return true
}

func deleteProject() bool {
	if activeProject != "" {

		req, err := http.NewRequest(http.MethodDelete, "/projects/"+activeProject, strings.NewReader(""))
		if err != nil {
			setStatus("Failed to delete "+activeProject, "bg-danger")
			return false
		}

		client := &http.Client{}

		response, err := client.Do(req)
		if err == nil {
			if response.StatusCode == http.StatusOK {
				setStatus("Succesfully deleted "+activeProject, "bg-success")
			} else {
				setStatus("Failed to delete "+activeProject, "bg-danger")
				return false
			}
		} else {
			setStatus("Failed to delete "+activeProject, "bg-danger")
			return false
		}

		jQuery("#project" + activeProject).Remove()

		setActiveProject(nil)
	}

	return true
}

func loadProjectCode(title string) {
	if title != "" {
		code, err := getProjectCode(title)
		if err == nil {
			js.Global.Get("editor").Call("setValue", code, -1)
		}
	} else {
		js.Global.Get("editor").Call("setValue", "", -1)
	}
}

func saveProjectCode() bool {
	if !IsProjects() {
		return true
	}

	if activeProject != "" {
		val := js.Global.Get("editor").Call("getValue")

		reader := strings.NewReader(val.String())

		req, err := http.NewRequest(http.MethodPut, "/projects/"+activeProject, reader)
		if err != nil {
			setStatus("Failed to save "+activeProject, "bg-danger")
			return false
		}

		req.Header.Set("Content-Type", "application/text")

		client := &http.Client{}

		response, err := client.Do(req)

		defer response.Body.Close()
		_, err = ioutil.ReadAll(response.Body)
		if err == nil {
			if response.StatusCode == http.StatusOK {
				setStatus("Succesfully saved "+activeProject, "bg-success")
			} else {
				setStatus("Failed to save "+activeProject, "bg-danger")
				return false
			}
		} else {
			setStatus("Failed to save "+activeProject, "bg-danger")
			return false
		}
	}

	return true
}

func getProjects() ([]string, error) {
	response, err := http.Get("/projects")
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if response.StatusCode == http.StatusOK {
		projects := []struct {
			Title string `json:"title"`
		}{}

		err = json.Unmarshal(data, &projects)
		if err != nil {
			return nil, err
		}

		res := make([]string, len(projects))
		for i, p := range projects {
			res[i] = p.Title
		}

		return res, nil

	} else {
		return nil, errors.New(string(data))
	}
}

func getProjectCode(title string) (string, error) {
	response, err := http.Get("/projects/" + title)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if response.StatusCode == http.StatusOK {
		project := struct {
			Title  string `json:"title"`
			Source string `json:"source"`
		}{}

		err = json.Unmarshal(data, &project)
		if err != nil {
			return "", err
		}

		return project.Source, nil

	} else {
		return "", errors.New(string(data))
	}
}

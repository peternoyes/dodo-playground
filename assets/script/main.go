// +build js

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/peternoyes/dodo-sim"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"net/http"
	"strings"
)

var jQuery = jquery.NewJQuery

var ctx *js.Object
var fram []byte
var firmware []byte
var stop bool
var speaker *WebSpeaker
var language string

func main() {
	loadAPI()
	projectsLogic()

	s := new(dodosim.SimulatorSync)
	stop = false

	speaker = new(WebSpeaker)
	speaker.New()

	language = "c"

	c := js.Global.Get("gameCanvas")
	ctx = c.Call("getContext", "2d")
	c.Set("width", 256)
	c.Set("height", 128)

	fram, _ = getAsset("fram.bin")
	firmware, _ = getAsset("firmware100.bin")

	wr := new(WebRenderer)
	wr.New(ctx)
	s.Renderer = wr

	fmt.Println("Initializing Speaker...")
	s.Speaker = speaker

	s.SimulateSyncInit(firmware, fram)

	// Load Code
	if !IsProjects() {
		id := getUrlParameter("code")
		if id != "" {
			code, lang, err := downloadCode(id)
			if err != nil {
				setStatus("Error Fetching Source: "+err.Error(), "bg-danger")
				return
			} else {
				js.Global.Get("editor").Call("setValue", code, -1)
				language = lang
				refreshLanguageDropdown()
			}
		} else {
			// Download Sample Application
			raw, _ := getAsset("sample.c")
			js.Global.Get("editor").Call("setValue", string(raw), -1)

		}
	}

	languageLogic()
	loginLogic()
	logoutLogic()
	flashLogic()
	runLogic(s)

	setStatus("Ready. Click 'Run' to try your game in the simulator.", "bg-success")
}

func refreshLanguageDropdown() {
	fmt.Println("Refreshing: ", language)
	switch language {
	case "c":
		jQuery("#activeLanguage").SetHtml("C")
		break
	case "assembly":
		jQuery("#activeLanguage").SetHtml("Assembly")
		break
	}
}

func languageLogic() {
	click := js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		go func() {
			language = strings.ToLower(jQuery(this).Text())
			jQuery("#activeLanguage").SetHtml(jQuery(this).Html())
		}()
		return nil
	})

	jQuery("#dropdownMenuLanguage").On(jquery.CLICK, "li a", click)
}

func loginLogic() {
	url := "/login"
	jQuery("#loginButton").On(jquery.CLICK, func() {
		go func() {

			/*
				newWindow := js.Global.Get("window").Call("open", url, "name", "height=600,width=450")
				if js.Global.Get("window").Get("focus") != js.Undefined {
					newWindow.Call("focus")
				}*/

			js.Global.Get("window").Get("location").Set("href", url)
		}()
	})
}

func logoutLogic() {
	url := "/logout"
	jQuery("#logoutButton").On(jquery.CLICK, func() {
		go func() {
			js.Global.Get("window").Get("location").Set("href", url)
		}()
	})
}

func getUrlParameter(param string) string {
	search := js.Global.Get("window").Get("location").Get("search").String()[1:]
	pageUrl := js.Global.Call("decodeURIComponent", search).String()

	vars := strings.Split(pageUrl, "&")
	for _, v := range vars {
		token := strings.Split(v, "=")
		if token[0] == param {
			return token[1]
		}
	}
	return ""
}

func compileCode() ([]byte, string, error) {
	val := js.Global.Get("editor").Call("getValue")

	reader := strings.NewReader(val.String())

	req, err := http.NewRequest(http.MethodPost, "/build", reader)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Content-Type", "application/text")
	req.Header.Set("X-Language", language)

	client := &http.Client{}

	response, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)

	if response.StatusCode == http.StatusOK {
		res := struct {
			Binary []byte `json:"binary"`
			Id     string `json:"id"`
		}{}

		err = json.Unmarshal(data, &res)

		fmt.Println(res.Binary)

		return res.Binary, res.Id, err
	} else {
		res := struct {
			Message string `json:"message"`
		}{}

		err = json.Unmarshal(data, &res)

		return nil, "", errors.New(res.Message)
	}
}

func downloadCode(id string) (string, string, error) {
	response, err := http.Get("/code/" + id)
	if err != nil {
		return "", "c", err
	}

	l := response.Header.Get("X-Language")

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", "c", err
	} else {
		return string(data), l, nil
	}
}

func getAsset(name string) ([]byte, error) {
	response, err := http.Get("/assets/" + name)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)

	return data, err
}

func setStatus(val, class string) {
	jQuery("#results").SetHtml(prepareTextForHTML(val))
	parent := js.Global.Get("results").Get("parentNode")
	setBgClass(parent, class)
}

func prepareTextForHTML(s string) string {
	return "<p>" + strings.Replace(s, "\n", "<br>", -1) + "</p>"
}

func setBgClass(j *js.Object, class string) {
	s := j.Get("className").String()

	tokens := strings.Split(s, " ")
	newClass := ""

	for _, t := range tokens {
		if !strings.HasPrefix(t, "bg") {
			if len(newClass) > 0 {
				newClass += " "
			}
			newClass += t
		}
	}

	if len(newClass) > 0 {
		newClass += " "
	}

	newClass += class

	j.Set("className", newClass)
}

func loadAPI() {
	api, _ := getAsset("api.md")
	output := blackfriday.MarkdownCommon(api)
	jQuery("#api").SetHtml(string(output))
}

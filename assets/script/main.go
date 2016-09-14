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
	"strconv"
	"strings"
	"time"
)

var jQuery = jquery.NewJQuery

var ctx *js.Object
var fram []byte
var firmware []byte
var stop bool
var speaker *WebSpeaker

func main() {
	loadAPI()

	s := new(dodosim.SimulatorSync)
	stop = false

	speaker = new(WebSpeaker)
	speaker.New()

	jQuery("#runButton").On(jquery.CLICK, func() {
		go func() {
			fmt.Println("Compile Initiated...")

			setStatus("Compiling...", "bg-info")

			data, id, err := compileCode()
			if err != nil {
				setStatus(err.Error(), "bg-danger")
				return
			}

			setStatus("Loading Simulator...", "bg-success")

			fram = data

			s.SwitchFram(fram)

			gamelink := js.Global.Get("document").Get("location").Get("origin").String()
			gamelink += "/?code="
			gamelink += id

			jQuery("#gamelink").SetText(gamelink)

			jQuery("#simModal").Call("modal")

			runSimulator(s)

			setStatus("Success! Simulator Running.", "bg-success")
		}()
	})

	jQuery("#simModal").On("hidden.bs.modal", func() {
		go func() {
			fmt.Println("About to stop simulator")
			stopSimulator()
			setStatus("Simulator Stopped.", "bg-success")
		}()
	})

	jQuery("#resetButton").On(jquery.CLICK, func() {
		go func() {
			s.Cpu.Reset(s.Bus)
		}()
	})

	jQuery("#muteButton").On(jquery.CLICK, func() {
		go func() {
			speaker.ToggleMute()
			jQuery("#muteButton").ToggleClass("active")
		}()
	})

	c := js.Global.Get("gameCanvas")
	ctx = c.Call("getContext", "2d")
	c.Set("width", 256)
	c.Set("height", 128)

	fram, _ = getAsset("fram.bin")
	firmware, _ = getAsset("firmware")

	wr := new(WebRenderer)
	wr.New(ctx)
	s.Renderer = wr

	fmt.Println("Initializing Speaker...")
	s.Speaker = speaker

	s.CyclesPerFrame = func(cycles uint64) {
		jQuery("#cycles").SetText(strconv.Itoa(int(cycles)))
	}

	s.SimulateSyncInit(firmware, fram)

	id := getUrlParameter("code")
	if id != "" {
		code, err := downloadCode(id)
		if err != nil {
			setStatus("Error Fetching Source: "+err.Error(), "bg-danger")
			return
		} else {
			js.Global.Get("editor").Call("setValue", code, -1)
		}
	} else {
		// Download Sample Application
		raw, _ := getAsset("sample.c")
		js.Global.Get("editor").Call("setValue", string(raw), -1)

	}

	setStatus("Ready. Click 'Run' to try your game in the simulator.", "bg-success")
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

func runSimulator(s *dodosim.SimulatorSync) {
	stop = false

	speaker.Enable()

	keyState := make(map[int]bool)

	downHandler := func(event *js.Object) {
		k := event.Get("keyCode").Int()
		event.Call("preventDefault")
		keyState[k] = true
	}

	upHandler := func(event *js.Object) {
		k := event.Get("keyCode").Int()
		event.Call("preventDefault")
		keyState[k] = false
	}

	js.Global.Call("addEventListener", "keydown", downHandler)
	js.Global.Call("addEventListener", "keyup", upHandler)

	// Measure 2 seconds worth, to figure out delay
	start := time.Now()
	for i := 0; i < 40; i++ {
		s.PumpClock("")
	}
	elapsed := time.Since(start)
	fmt.Println(elapsed)

	// Calculate the delay necessary to achieve 20FPS
	delay := 1.0
	seconds := elapsed.Seconds()
	if seconds < 1.5 {
		delay = ((2.0 - seconds) / 40.0) * 1000.0
	}

	fmt.Println("Delay of :", delay)

	Every(time.Millisecond*time.Duration(delay), func() bool {
		k := ""
		if keyState[37] {
			k += "L"
		}
		if keyState[38] {
			k += "U"
		}
		if keyState[39] {
			k += "R"
		}
		if keyState[40] {
			k += "D"
		}
		if keyState[65] {
			k += "A"
		}
		if keyState[66] {
			k += "B"
		}

		s.PumpClock(k)

		if stop {
			js.Global.Call("removeEventListener", "keydown", downHandler)
			js.Global.Call("removeEventListener", "keyup", upHandler)
		}

		return !stop
	})
}

func stopSimulator() {
	speaker.Disable()
	stop = true
}

func Every(duration time.Duration, fn func() bool) {
	time.AfterFunc(duration, func() {
		if !fn() {
			return
		}
		Every(duration, fn)
	})
}

func compileCode() ([]byte, string, error) {
	val := js.Global.Get("editor").Call("getValue")

	reader := strings.NewReader(val.String())

	response, err := http.Post("/build", "application/text", reader)

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)

	if response.StatusCode == http.StatusOK {
		res := struct {
			Binary []byte `json:"binary"`
			Id     string `json:"id"`
		}{}

		err = json.Unmarshal(data, &res)

		return res.Binary, res.Id, err
	} else {
		res := struct {
			Message string `json:"message"`
		}{}

		err = json.Unmarshal(data, &res)

		return nil, "", errors.New(res.Message)
	}
}

func downloadCode(id string) (string, error) {
	response, err := http.Get("/code/" + id)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	} else {
		return string(data), nil
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

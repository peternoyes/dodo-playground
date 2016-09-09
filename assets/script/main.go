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
var on *js.Object
var off *js.Object
var fram []byte
var firmware []byte
var stop bool

func main() {
	loadAPI()

	s := new(dodosim.SimulatorSync)
	stop = false

	jQuery("#runButton").On(jquery.CLICK, func() {
		go func() {
			fmt.Println("Compile Initiated...")

			setStatus("Compiling...", "bg-info")

			data, err := compileCode()
			if err != nil {
				setStatus(err.Error(), "bg-danger")
				return
			}

			setStatus("Loading Simulator...", "bg-success")

			fram = data

			s.SwitchFram(fram)

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

	c := js.Global.Get("gameCanvas")
	ctx = c.Call("getContext", "2d")
	c.Set("width", 256)
	c.Set("height", 128)

	on = ctx.Call("createImageData", 2, 2)
	data := on.Get("data")
	for i := 0; i < 4; i++ {
		data.SetIndex(i*4+0, 255)
		data.SetIndex(i*4+1, 255)
		data.SetIndex(i*4+2, 255)
		data.SetIndex(i*4+3, 255)
	}

	off = ctx.Call("createImageData", 2, 2)
	data = off.Get("data")
	for i := 0; i < 4; i++ {
		data.SetIndex(i*4+0, 0)
		data.SetIndex(i*4+1, 0)
		data.SetIndex(i*4+2, 0)
		data.SetIndex(i*4+3, 255)
	}

	fram, _ = getAsset("fram.bin")
	firmware, _ = getAsset("firmware")

	s.Renderer = new(WebRenderer)

	s.CyclesPerFrame = func(cycles uint64) {
		jQuery("#cycles").SetText(strconv.Itoa(int(cycles)))
	}

	s.SimulateSyncInit(firmware, fram)

	setStatus("Ready. Click 'Run' to try your game in the simulator.", "bg-success")
}

func runSimulator(s *dodosim.SimulatorSync) {
	stop = false

	keyState := make(map[int]bool)

	js.Global.Call("addEventListener", "keydown", func(event *js.Object) {
		k := event.Get("keyCode").Int()
		keyState[k] = true
	})

	js.Global.Call("addEventListener", "keyup", func(event *js.Object) {
		k := event.Get("keyCode").Int()
		keyState[k] = false
	})

	// Measure 2 seconds worth, to figure out delay
	start := time.Now()
	for i := 0; i < 40; i++ {
		s.PumpClock("")
	}
	elapsed := time.Since(start)
	fmt.Println(elapsed)

	Every(time.Millisecond*1, func() bool {
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

		return !stop
	})
}

func stopSimulator() {
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

type WebRenderer struct {
}

func (r *WebRenderer) Render(data [1024]byte) {
	//fmt.Println("Render Called")
	var x, y int
	var b byte
	for y = 0; y < 64; y++ {
		for x = 0; x < 128; x++ {
			b = (data[x+((y/8)*128)] >> (byte(y) % 8)) & 1
			if b == 1 {
				ctx.Call("putImageData", on, x*2, y*2)
			} else {
				ctx.Call("putImageData", off, x*2, y*2)
			}
		}
	}
}

func compileCode() ([]byte, error) {
	val := js.Global.Get("editor").Call("getValue")

	reader := strings.NewReader(val.String())

	response, err := http.Post("/build", "application/text", reader)

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)

	if response.StatusCode == http.StatusOK {
		res := struct {
			Binary []byte `json:"binary"`
		}{}

		err = json.Unmarshal(data, &res)

		return res.Binary, err
	} else {
		res := struct {
			Message string `json:"message"`
		}{}

		err = json.Unmarshal(data, &res)

		return nil, errors.New(res.Message)
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

	js.Global.Get("api").Set("innerHTML", string(output))
}

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/peternoyes/dodo-sim"
)

func getUsedSpace(data []byte) int {
	for i := 8191; i >= 0; i-- {
		if data[i] != 0 {
			return (i * 100) / 8192
		}
	}

	return 0
}

func runLogic(s *dodosim.SimulatorSync) {
	jQuery("#runButton").On(jquery.CLICK, func() {
		go func() {
			if !saveProjectCode() {
				return
			}

			fmt.Println("Compile Initiated...")

			setStatus("Compiling...", "bg-info")

			data, id, err := compileCode()
			if err != nil {
				setStatus(err.Error(), "bg-danger")
				return
			}

			usedSpace := getUsedSpace(data)

			setStatus("Loading Simulator...", "bg-success")

			fram = data

			if activeProject != "" {
				item := localStorage.Call("getItem", activeProject)
				if item != nil && item != js.Undefined {
					userData := make([]uint8, 64)
					err := json.Unmarshal([]byte("["+item.String()+"]"), &userData)

					for i := 0; i < 64; i++ {
						fram[i+4] = userData[i]
					}
				}
			}

			firmwareBytes, err := getFirmware(version)
			if err != nil {
				setStatus(err.Error(), "bg-danger")
				return
			}

			s.SwitchFram(firmwareBytes, fram)

			gamelink := js.Global.Get("document").Get("location").Get("origin").String()
			gamelink += "/?code="
			gamelink += id

			jQuery("#gamelink").SetText(gamelink)

			jQuery("#simModal").Call("modal")

			runSimulator(s)

			setStatus("Success! Simulator Running.\nUsed Space: "+strconv.Itoa(usedSpace)+"%", "bg-success")
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

	jQuery("#copyButton").On(jquery.CLICK, func() {
		go func() {
			parent := jQuery("#simModal")
			temp := jQuery(js.Global.Get("document").Call("createElement", "input"))
			parent.Call("append", temp)
			temp.SetVal(jQuery("#gamelink").Text()).Select()
			js.Global.Get("document").Call("execCommand", "copy")
			temp.Remove()
		}()
	})

	jQuery("#muteButton").On(jquery.CLICK, func() {
		go func() {
			speaker.ToggleMute()
			jQuery("#muteButton").ToggleClass("active")
		}()
	})

	s.CyclesPerFrame = func(cycles uint64) {
		jQuery("#cycles").SetText(strconv.Itoa(int(cycles)))
	}
}

func runSimulator(s *dodosim.SimulatorSync) {
	stop = false

	speaker.Enable()

	keyState := make(map[int]bool)
	buttonState := make(map[string]bool)
	buttons := []string{"upButton", "leftButton", "rightButton", "downButton", "aButton", "bButton"}

	downHandler := func(event *js.Object) {
		k := event.Get("keyCode").Int()
		if event.Get("ctrlKey").Bool() || k == 91 || k == 67 { // Allow Ctrl+C
			return
		}

		event.Call("preventDefault")
		keyState[k] = true
	}

	upHandler := func(event *js.Object) {
		k := event.Get("keyCode").Int()
		if event.Get("ctrlKey").Bool() || k == 91 || k == 67 {
			return
		}

		event.Call("preventDefault")
		keyState[k] = false
	}

	buttonDownHandler := func(event *js.Object) {
		go func() {
			event.Call("preventDefault")
			tgt := event.Get("target")
			if tgt.Get("tagName").String() == "SPAN" {
				tgt = tgt.Get("parentNode")
			}

			id := tgt.Get("id").String()
			buttonState[id] = true
		}()
	}

	buttonUpHandler := func(event *js.Object) {
		go func() {
			tgt := event.Get("target")
			if tgt.Get("tagName").String() == "SPAN" {
				tgt = tgt.Get("parentNode")
			}

			id := tgt.Get("id").String()
			buttonState[id] = false
		}()
	}

	js.Global.Call("addEventListener", "keydown", downHandler)
	js.Global.Call("addEventListener", "keyup", upHandler)

	for _, s := range buttons {
		jQuery("#"+s).On(jquery.MOUSEDOWN, func(event *js.Object) { buttonDownHandler(event) })
		jQuery("#"+s).On(jquery.MOUSEUP, func(event *js.Object) { buttonUpHandler(event) })

		jQuery("#"+s).On(jquery.TOUCHSTART, func(event *js.Object) { buttonDownHandler(event) })
		jQuery("#"+s).On(jquery.TOUCHEND, func(event *js.Object) { buttonUpHandler(event) })
	}

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
		if keyState[37] || buttonState["leftButton"] {
			k += "L"
		}
		if keyState[38] || buttonState["upButton"] {
			k += "U"
		}
		if keyState[39] || buttonState["rightButton"] {
			k += "R"
		}
		if keyState[40] || buttonState["downButton"] {
			k += "D"
		}
		if keyState[65] || buttonState["aButton"] {
			k += "A"
		}
		if keyState[66] || buttonState["bButton"] {
			k += "B"
		}

		s.PumpClock(k)

		if stop {
			js.Global.Call("removeEventListener", "keydown", downHandler)
			js.Global.Call("removeEventListener", "keyup", upHandler)

			for _, s := range buttons {
				jQuery("#"+s).Off(jquery.MOUSEDOWN, nil)
				jQuery("#"+s).Off(jquery.MOUSEUP, nil)
			}
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

package main

import (
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/peternoyes/dodo-sim"
	"strconv"
	"time"
)

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

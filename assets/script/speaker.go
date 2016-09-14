// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
)

type WebSpeaker struct {
	Context    *js.Object
	Oscillator *js.Object
	Gain       *js.Object
}

func (s *WebSpeaker) New() {
	v := js.Global.Get("window").Get("AudioContext")
	if v == js.Undefined {
		v = js.Global.Get("window").Get("webkitAudioContext")
	}

	s.Context = v.New()
	s.Gain = s.Context.Call("createGain")
}

func (s *WebSpeaker) SetFrequency(freq int) {
	if freq == 0 {
		s.Stop()
		return
	}

	if s.Oscillator != nil {
		s.Oscillator.Call("stop")
	}

	s.Oscillator = s.Context.Call("createOscillator")
	s.Oscillator.Set("type", "square")
	s.Oscillator.Get("frequency").Set("value", freq)
	s.Gain.Get("gain").Set("value", 0.05)

	s.Oscillator.Call("connect", s.Gain)
	s.Gain.Call("connect", s.Context.Get("destination"))

	s.Oscillator.Call("start")
}

func (s *WebSpeaker) Stop() {
	s.Gain.Get("gain").Set("value", 0)
}

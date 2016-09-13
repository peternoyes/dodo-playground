// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
)

type WebRenderer struct {
	Context *js.Object
	Blocks  []*js.Object
}

func (r *WebRenderer) New(ctx *js.Object) {
	r.Context = ctx
	r.Blocks = make([]*js.Object, 256)
	mask := []byte{1, 2, 4, 8, 16, 32, 64, 128}
	for i := 0; i < 256; i++ {
		n := byte(i)
		block := ctx.Call("createImageData", 2, 16)
		data := block.Get("data")
		r.Blocks[i] = block
		for y := 0; y < 8; y++ {
			r := 0
			g := 0
			b := 0
			if mask[y]&n != 0 {
				g = 255
			}

			for yc := y * 2; yc < (y*2)+2; yc++ {
				for xc := 0; xc < 2; xc++ {
					o := yc*2*4 + xc*4
					data.SetIndex(o+0, r)
					data.SetIndex(o+1, g)
					data.SetIndex(o+2, b)
					data.SetIndex(o+3, 255)
				}
			}
		}
	}
}

func (r *WebRenderer) Render(data [1024]byte) {
	for p := 0; p < 8; p++ {
		for x := 0; x < 128; x++ {
			b := data[p*128+x]
			xc := x * 2
			yc := p * 16
			r.Context.Call("putImageData", r.Blocks[b], xc, yc)
		}
	}
}

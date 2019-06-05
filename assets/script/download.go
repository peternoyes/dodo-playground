package main

import (
	"github.com/flimzy/jsblob"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

func downloadLogic() {
	jQuery("#downloadButton").On(jquery.CLICK, func() {
		go func() {
			setStatus("Compiling...", "bg-info")

			var err error
			data, _, err := compileCode()
			if err != nil {
				setStatus(err.Error(), "bg-danger")
				return
			}

			buffer := js.NewArrayBuffer(data)
			blob := jsblob.New([]interface{}{buffer}, jsblob.Options{Type: "application/octet-string"})

			link := js.Global.Get("document").Call("createElement", "a")
			link.Set("href", js.Global.Get("window").Get("URL").Call("createObjectURL", blob))

			link.Set("download", "fram.bin")
			link.Call("click")

			setStatus("Successfully created fram.bin", "bg-success")
		}()
	})
}

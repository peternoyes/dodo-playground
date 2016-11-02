package main

import (
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"strconv"
)

type Port struct {
	*js.Object
	Name string `js:"name"`
}

type Message struct {
	*js.Object
	Message string `js:"message"`
}

type Fram struct {
	*js.Object
	Fram []byte `js:"fram"`
	Path string `js:"path"`
}

func flashLogic() {
	var data []byte
	activePort := ""

	chrome := js.Global.Get("chrome")
	if chrome != nil && chrome != js.Undefined {
		msg := Message{Object: js.Global.Get("Object").New()}
		msg.Message = "version"

		js.Global.Get("chrome").Get("runtime").Call("sendMessage", "bckholjcbphjhdfgbejkjflcafdgbdkb", msg, func(reply *js.Object) {
			go func() {
				if reply != nil && reply != js.Undefined {
					v := reply.Get("version").Float()
					if v >= 1.0 {
						jQuery("#flashButton").Show()
					}
				}
			}()
		})
	}

	jQuery("#flashBeginButton").On(jquery.CLICK, func() {
		go func() {
			if activePort != "" && data != nil {

				fmt.Println("About to connect")

				jQuery("#dropdownMenuPorts").SetProp("disabled", true)
				jQuery("#flashBeginButton").SetProp("disabled", true)

				portObj := Port{Object: js.Global.Get("Object").New()}
				portObj.Name = "dodo_flash"

				port := js.Global.Get("chrome").Get("runtime").Call("connect", "bckholjcbphjhdfgbejkjflcafdgbdkb", portObj)

				fmt.Println(port)

				dataObj := Fram{Object: js.Global.Get("Object").New()}
				dataObj.Fram = data
				dataObj.Path = activePort

				port.Call("postMessage", dataObj)

				port.Get("onMessage").Call("addListener", func(reply *js.Object) {
					go func() {
						if reply != nil && reply != js.Undefined {
							progressObj := reply.Get("progress")
							if progressObj != nil && progressObj != js.Undefined {
								progress := progressObj.Int()
								jQuery("#flashprogress").SetCss("width", strconv.Itoa(progress)+"%").SetAttr("aria-valuenow", progress)
							} else {
								successObj := reply.Get("success")
								if successObj != nil && successObj != js.Undefined {
									success := successObj.Bool()
									if success {

									} else {

									}
								} else {
									// Check Error Obj
								}
							}
						}
					}()
				})
			}
		}()
	})

	click := js.MakeFunc(func(this *js.Object, arguments []*js.Object) interface{} {
		go func() {
			activePort = jQuery(this).Text()
			jQuery("#activePort").SetHtml(jQuery(this).Html())
			jQuery("#flashBeginButton").SetProp("disabled", false)
		}()
		return nil
	})

	jQuery("#dropdownMenuPortsItems").On(jquery.CLICK, "li a", click)

	jQuery("#flashButton").On(jquery.CLICK, func() {
		go func() {
			if !saveProjectCode() {
				return
			}

			activePort = ""

			jQuery("#flashBeginButton").SetProp("disabled", true)

			setStatus("Compiling...", "bg-info")

			var err error
			data, _, err = compileCode()
			if err != nil {
				setStatus(err.Error(), "bg-danger")
				return
			}

			setStatus("Flashing...", "bg-info")

			jQuery("#dropdownMenuPorts").SetProp("disabled", false)
			jQuery("#dropdownMenuPortsItems").Children("").Remove()
			jQuery("activePort").SetHtml("Select a COM Port")
			jQuery("#flashprogress").SetCss("width", "0%").SetAttr("aria-valuenow", 0)

			msg := Message{Object: js.Global.Get("Object").New()}
			msg.Message = "devices"

			js.Global.Get("chrome").Get("runtime").Call("sendMessage", "bckholjcbphjhdfgbejkjflcafdgbdkb", msg, func(reply *js.Object) {
				go func() {
					if reply != nil && reply != js.Undefined {
						l := reply.Length()
						for i := 0; i < l; i++ {
							port := reply.Index(i)
							path := port.Get("path").String()
							displayName := path

							obj := port.Get("displayName")
							if obj != js.Undefined {
								//displayName = obj.String()
							}

							jQuery("#dropdownMenuPortsItems").Append("<li><a href='#'>" + displayName + "</a></li>")
						}
					}
				}()
			})

			jQuery("#flashModal").Call("modal")

		}()
	})
}

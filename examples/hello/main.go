package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/kanryu/slint"
)

const ui = `
export component App inherits Window {
    width: 320px;
    height: 180px;
    in-out property <string> message: "Slint from Go";
    Text { text: message; horizontal-alignment: center; vertical-alignment: center; }
}
`

func main() {
	// Slint's host-side event pump must stay on one OS thread.  In
	// particular, do not let the Go scheduler move the callback/event
	// handling loop between native threads.
	runtime.LockOSThread()

	app, err := slint.New(ui)
	if err != nil {
		panic(err)
	}
	defer app.Close()
	if err := app.SetStringProperty("message", "SlintCFlat + Go"); err != nil {
		panic(err)
	}
	fmt.Println("Slint context created; running:", app.IsRunning())

	for app.IsRunning() {
		for {
			event, ok, err := app.PollEvent()
			if err != nil {
				panic(err)
			}
			if !ok {
				break
			}
			if err := app.DispatchEvent(event); err != nil {
				panic(err)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

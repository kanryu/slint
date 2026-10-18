package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/kanryu/slint"
)

func currentDirectoryFiles() (string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return "", err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, "\n"), nil
}

func main() {
	// Keep the host event pump on one native thread.  This is required when
	// callbacks/events cross the Rust/Go boundary on Windows.
	runtime.LockOSThread()

	files, err := currentDirectoryFiles()
	if err != nil {
		panic(err)
	}

	// ui_bundle.slint はMakefileのslint-bundlerターゲットで生成され、
	// go:embedによって起動時にこのstringへ読み込まれる。
	app, err := slint.New(bundledSlint)
	if err != nil {
		panic(err)
	}
	defer app.Close()

	if err := app.SetStringProperty("files", files); err != nil {
		panic(err)
	}
	fmt.Println("MVC router sample started:", app.IsRunning())

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

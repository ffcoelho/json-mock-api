package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ffcoelho/jma/keys"
)

var tty *keys.TTY

var mockDelay int
var mockStatus int = http.StatusOK
var mockStatusIdx int
var codes []int = []int{200, 201, 204, 301, 302, 304, 400, 401, 403, 404, 409, 410, 500, 501}
var lastPrint int = -1

func main() {
	setupExitHandler()
	setupKeyboardHandler()
	http.HandleFunc("/", handler)
	fmt.Printf("JSON Mock API is up.\nListening on http://localhost:%d\n", 9000)
	http.ListenAndServe(":9000", nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	br := ""
	if lastPrint != 0 {
		lastPrint = 0
		br = "\n"
	}

	now := time.Now()
	fmt.Printf("%s%s %d %s %s", br, now.Format("15:04:05"), mockStatus, req.Method, req.URL.Path)
	w.WriteHeader(mockStatus)
	w.Header().Set("Content-Type", "application/json")
	time.Sleep(time.Duration(mockDelay) * time.Millisecond)
	w.Write([]byte("Ok"))
}

func toggleDelay() {
	br := ""
	if lastPrint != 1 {
		lastPrint = 1
		br = "\n"
	}
	fmt.Printf("%s\r                  ", br)
	switch mockDelay {
	case 0:
		mockDelay = 600
	case 600:
		mockDelay = 2000
	case 2000:
		mockDelay = 5000
	default:
		mockDelay = 0
		fmt.Printf("\rMock delay: OFF")
		return
	}
	fmt.Printf("\rMock delay: %dms", mockDelay)
}

func toggleStatus(i int) {
	br := ""
	if lastPrint != 2 {
		lastPrint = 2
		br = "\n"
	}

	mockStatusIdx += i
	if mockStatusIdx == -1 {
		mockStatusIdx = len(codes) - 1
	}
	if mockStatusIdx == len(codes) {
		mockStatusIdx = 0
	}
	mockStatus = codes[mockStatusIdx]
	fmt.Printf("%s\rMock status code: %d", br, codes[mockStatusIdx])
}

func setupExitHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\rJSON Mock API is down.")
		tty.Close()
		os.Exit(0)
	}()
}

func setupKeyboardHandler() {
	go func() {
		tty = keys.Open()
		if tty == nil {
			return
		}

		for {
			key := tty.ReadKey()
			switch key {
			case "a":
				toggleStatus(-1)
			case "d":
				toggleDelay()
			case "s":
				toggleStatus(1)
			}
		}
	}()
}

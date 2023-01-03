package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ffcoelho/jma/keys"
)

var tty *keys.TTY

var delay int
var statusCode int = 1
var statusCodeIdx int

func setupExitHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		tty.Close()
		os.Exit(0)
	}()
}

func setupKeyboardHandler() {
	go func() {
		tty = keys.Open()
		for {
			key := tty.ReadKey()
			switch key {
			case "a":
				toggleStatus(-1)
			case "s":
				toggleStatus(1)
			case "d":
				toggleDelay()
			}
		}
	}()
}

func toggleDelay() {
	br := ""
	if lastPrint != 1 {
		lastPrint = 1
		br = "\n"
	}
	fmt.Printf("%s\r              ", br)
	switch delay {
	case 0:
		delay = 600
	case 600:
		delay = 1200
	case 1200:
		delay = 2400
	default:
		delay = 0
		fmt.Printf("\rDELAY: off")
		return
	}
	fmt.Printf("\rDELAY: %dms", delay)
}

func toggleStatus(i int) {
	br := ""
	if lastPrint != 2 {
		lastPrint = 2
		br = "\n"
	}
	statusCodeIdx += i
	if statusCodeIdx == -1 {
		statusCodeIdx = len(codes) - 1
	} else if statusCodeIdx == len(codes) {
		statusCodeIdx = 0
	}
	statusCode = codes[statusCodeIdx]
	if statusCode == 1 {
		fmt.Printf("%s\rSTATUS CODE: 2xx", br)
	} else {
		fmt.Printf("%s\rSTATUS CODE: %d", br, codes[statusCodeIdx])
	}
}

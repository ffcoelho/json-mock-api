package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/ffcoelho/jma/keys"
)

func main() {
	setupExitHandler()
	setupKeyboardHandler()
	http.HandleFunc("/", handler)
	fmt.Printf("JSON Mock API is up.\nListening on http://localhost:%d\n", 9000)
	http.ListenAndServe(":9000", nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Ok"))
}

func setupExitHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\rJSON Mock API is down.")
		exec.Command("stty", "-F", "/dev/tty", "echo").Run()
		os.Exit(0)
	}()
}

func setupKeyboardHandler() {
	go func() {
		t, err := keys.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer t.Close()

		for {
			key := t.ReadKey()
			if key != "d" && key != "s" {
				continue
			}
			fmt.Printf("KEY: %s\n", key)
		}
	}()
}

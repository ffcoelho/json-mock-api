package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	setupExitHandler()
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
		os.Exit(0)
	}()
}

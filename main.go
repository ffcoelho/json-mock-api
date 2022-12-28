package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	fmt.Printf("Listening on http://localhost:%d\n", 9000)
	http.ListenAndServe(":9000", nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Ok"))
}

package main

import (
	"fmt"
	"net/http"
)

func startServer1() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Response from server 1")
    })
    http.ListenAndServe(":8081", nil)	
}

func main() {
	startServer1()
}

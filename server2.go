package main

import (
	"fmt"
	"net/http"
)

func startServer2() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Response from server 2")
    })
    http.ListenAndServe(":8082", nil)	
}

func main() {
	startServer2()
}

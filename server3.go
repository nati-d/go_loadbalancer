package main

import (
	"fmt"
	"net/http"
)

func startServer3() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from server 3")
	})
	http.ListenAndServe(":8083", nil)
}

func main() {
	startServer3()
}

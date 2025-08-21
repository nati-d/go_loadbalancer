package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)


var servers = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
}

var currentServer = 0


// round robin algorithm picks the next backend in a circular way
func roundRobin() string {
	server := servers[currentServer]
	currentServer = (currentServer + 1) % len(servers)
	return server
}


//handler forwards the request to a backend server
func handler(w http.ResponseWriter, r *http.Request) {
	server := roundRobin()

	//Parse the server URL
	url, err := url.Parse(server)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Create a reverse proxy
	req, err := http.NewRequest(r.Method, url.String() + r.RequestURI, r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Copy the headers from the original request
	req.Header = r.Header

	//Send the request to the backend server
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	//Copy the response headers from the backend server to the client
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

//Go routines to start all the servers
func startServer(port string, message string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, message)
	})

	server := &http.Server{
		Addr: port,
		Handler: mux,
	}

	log.Printf("Server %s started on port %s", message, port)
	log.Fatal(server.ListenAndServe())
}

func main() {
    // Start backend servers
    go startServer(":8081", "Response from Server 1")
    go startServer(":8082", "Response from Server 2")
    go startServer(":8083", "Response from Server 3")

    // Start load balancer
    fmt.Println("Starting load balancer... at port 8080")
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))

    select {} 
}
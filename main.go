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
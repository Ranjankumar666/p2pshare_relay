package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func CreateProxy() {
	RELAY_PORT, ok := os.LookupEnv("RELAY_PORT")

	if !ok {
		RELAY_PORT = "8080"
	}

	PORT, ok := os.LookupEnv("PORT")

	if !ok {
		PORT = "4000"
	}

	targetURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%s", RELAY_PORT))

	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// WebSocket support
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
	}

	http.HandleFunc("/", proxy.ServeHTTP)

	log.Printf("Reverse proxy listening on :%s\n", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", PORT), nil))
}

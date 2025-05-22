package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func CreateProxy(ctx context.Context) {
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

	handlers := http.NewServeMux()

	handlers.HandleFunc("/", proxy.ServeHTTP)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: handlers,
	}

	go func() {
		log.Printf("Reverse proxy listening on :%s\n", PORT)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Proxy server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down proxy")

	if err := server.Close(); err != nil {
		log.Printf("Error while closing proxy server: %s", err.Error())
	}

}

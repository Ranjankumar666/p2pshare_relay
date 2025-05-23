package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// go CreateProxy(ctx)
	go CreateServer(ctx)

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)

	<-sigChannel
	log.Printf("Shut down initiated...")
	cancel()

	time.Sleep(1 * time.Second)
	// select {}
}

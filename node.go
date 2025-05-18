package main

import (
	"context"
	"fmt"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/host/autonat"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
)

func CreateServer() {
	_, cancel := context.WithCancel(context.Background())

	defer cancel()
	PORT, ok := os.LookupEnv("RELAY_PORT")

	if !ok {
		PORT = "8080"
	}

	addresses := []string{
		// fmt.Sprintf("/ip4/127.0.0.1/tcp/%s/ws", PORT),
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%s/ws", PORT),
	}

	id, _ := LoadOrCreateIdentity()

	server, err := libp2p.New(
		libp2p.ListenAddrStrings(addresses...),
		libp2p.Identity(id),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.EnableRelay(),
		libp2p.EnableAutoNATv2(),
	)

	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}

	_, err = relay.New(server)
	if err != nil {
		log.Fatalf("Failed to start relay v2 service: %v", err)
	}

	identify.NewIDService(server)

	_, err = autonat.New(server)
	if err != nil {
		log.Fatalf("Failed to start AutoNAT: %v", err)
	}

	log.Println("Relay server is running at:")
	for _, addr := range server.Addrs() {
		log.Printf("%s/p2p/%s\n", addr, server.ID().String())
	}

	sub, _ := server.EventBus().Subscribe(new(event.EvtPeerConnectednessChanged))

	// listen to peer connect and disconnect
	go func() {
		for e := range sub.Out() {
			evt := e.(event.EvtPeerConnectednessChanged)
			log.Printf("Peer %s changed state: %s\n", evt.Peer, evt.Connectedness)
		}
	}()

	server.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			log.Printf("Connected: %s", conn.RemotePeer())
		},
		DisconnectedF: func(net network.Network, conn network.Conn) {
			log.Printf("Disconnected: %s", conn.RemotePeer())
		},
	})
}

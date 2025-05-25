package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"

	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
)

func CreateServer(ctx context.Context) {
	// _, cancel := context.WithCancel(context.Background())

	cgMgr, _ := connmgr.NewConnManager(200, 400, connmgr.WithGracePeriod(2*time.Minute))

	// defer cancel()
	PORT, ok := os.LookupEnv("RELAY_PORT")

	if !ok {
		PORT = "8080"
	}
	WSS_PORT, ok := os.LookupEnv("RELAY_WSS_PORT")

	if !ok {
		WSS_PORT = "443"
	}

	addresses := []string{
		// fmt.Sprintf("/ip4/127.0.0.1/tcp/%s/ws", PORT),
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%s/ws", PORT),
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%s/wss", WSS_PORT),
	}

	id, _ := LoadOrCreateIdentity()

	server, err := libp2p.New(
		libp2p.ConnectionManager(cgMgr),
		libp2p.ListenAddrStrings(addresses...),
		libp2p.Identity(id),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.EnableRelayService(relay.WithACL(&MyACLFilter{})),
		libp2p.EnableNATService(),
		libp2p.EnableAutoNATv2(),
		libp2p.EnableHolePunching(),
		libp2p.ForceReachabilityPublic(),
	)

	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}

	// Identity
	_, err = identify.NewIDService(server)

	if err != nil {
		log.Fatalf("Failed to start hole punching service: %v", err)
	}

	log.Println("Relay server is running at:")
	for _, addr := range server.Addrs() {
		log.Printf("%s/p2p/%s\n", addr, server.ID().String())
	}

	for _, p := range server.Mux().Protocols() {
		log.Printf("Supported protocol: %s", p)
	}

	server.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			log.Printf("Connected: %s", conn.RemotePeer())
			peers := server.Peerstore()

			log.Printf("\nConnected peers: %d", len(peers.Peers()))
			for id, peerID := range peers.Peers() {

				log.Printf("%d Peer ID: %s", id, peerID.String())
			}

		},
		DisconnectedF: func(net network.Network, conn network.Conn) {
			log.Printf("Disconnected: %s", conn.RemotePeer())

			peers := server.Peerstore()

			log.Printf("\nConnected peers: %d", len(peers.Peers()))
			for id, peerID := range peers.Peers() {
				log.Printf("%d Peer ID: %s", id, peerID.String())
			}
		},
	})

	<-ctx.Done()

	log.Println("Shutting down relay")

	if err := server.Close(); err != nil {
		log.Printf("Error while shutting relay :%v", err)
	}

}

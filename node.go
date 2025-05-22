package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/host/autonat"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"

	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	libp2pwebtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"
	"github.com/multiformats/go-multiaddr"
)

func CreateServer(ctx context.Context) {
	// _, cancel := context.WithCancel(context.Background())

	cgMgr, _ := connmgr.NewConnManager(200, 400, connmgr.WithGracePeriod(2*time.Minute))

	// defer cancel()
	PORT, ok := os.LookupEnv("RELAY_PORT")

	if !ok {
		PORT = "8080"
	}

	UDP_PORT, ok := os.LookupEnv("RELAY_UDP_PORT")

	if !ok {
		UDP_PORT = "9000"
	}

	addresses := []string{
		// fmt.Sprintf("/ip4/127.0.0.1/tcp/%s/ws", PORT),
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%s/ws", PORT),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%s/quic-v1/webtransport", UDP_PORT), // QUIC/WebTransport
	}

	id, _ := LoadOrCreateIdentity()

	server, err := libp2p.New(
		libp2p.ConnectionManager(cgMgr),
		libp2p.ListenAddrStrings(addresses...),
		libp2p.Identity(id),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New),
		libp2p.Transport(libp2pwebtransport.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.EnableAutoNATv2(),
		libp2p.EnableHolePunching(),
		libp2p.ForceReachabilityPrivate(),
		libp2p.ForceReachabilityPublic(),
	)

	if err != nil {
		log.Fatalf("Failed to create libp2p host: %v", err)
	}

	// Relay
	_, err = relay.New(server, relay.WithACL(&MyACLFilter{}))

	if err != nil {
		log.Fatalf("Failed to start relay v2 service: %v", err)
	}

	// Identity
	idService, _ := identify.NewIDService(server)

	_, err = autonat.New(server)
	if err != nil {
		log.Fatalf("Failed to start AutoNAT: %v", err)
	}

	// Holepunch
	_, err = holepunch.NewService(server, idService, func() []multiaddr.Multiaddr {
		log.Println("Getting listen addresses")

		return server.Addrs()
	})

	if err != nil {
		log.Fatalf("Failed to start hole punching service: %v", err)
	}

	log.Println("Relay server is running at:")
	for _, addr := range server.Addrs() {
		log.Printf("%s/p2p/%s\n", addr, server.ID().String())
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

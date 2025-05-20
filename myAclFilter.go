package main

import (
	"log"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type MyACLFilter struct{}

func (acl *MyACLFilter) AllowReserve(p peer.ID, a multiaddr.Multiaddr) bool {
	log.Printf("Incoming reservation : %s", p.String())
	return true
}
func (acl *MyACLFilter) AllowConnect(src peer.ID, srcAddr multiaddr.Multiaddr, dest peer.ID) bool {
	log.Printf("Connecting : %s to %s", src.String(), dest.String())

	return true
}

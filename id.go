package main

import (
	"crypto/rand"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func LoadOrCreateIdentity() (crypto.PrivKey, error) {
	const keyFile = "peerkey.key"

	if _, err := os.Stat(keyFile); err == nil {
		data, _ := os.ReadFile(keyFile)
		return crypto.UnmarshalPrivateKey(data)
	}

	priv, _, _ := crypto.GenerateEd25519Key(rand.Reader)
	data, _ := crypto.MarshalPrivateKey(priv)
	os.WriteFile(keyFile, data, 0600)
	return priv, nil
}

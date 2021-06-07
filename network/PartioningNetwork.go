package network

import (
	"crypto/rsa"
	"math/rand"
	"time"

	"github.com/Hvassaa/bachelor/signature"
	"github.com/Hvassaa/bachelor/types"
)

type PartioningNetwork struct {
	clients    []*ClientInterface
	partition1 map[string]struct{}
	partition2 map[string]struct{}
	delay      int
}

func (n *PartioningNetwork) Send(data interface{}) {
	for _, client := range n.clients {
		go n.Deliver(data, client)
	}
}

func MakePartioningNetwork(keys []rsa.PublicKey, delay int) Network {
	noOfKeys := len(keys)
	partition1 := make(map[string]struct{})
	partition2 := make(map[string]struct{})
	for i, key := range keys {
		if i < noOfKeys/2 {
			partition1[signature.PublicKeyToString(key)] = struct{}{}
		} else {
			partition2[signature.PublicKeyToString(key)] = struct{}{}
		}
	}
	network := PartioningNetwork{clients: make([]*ClientInterface, 0), partition1: partition1, partition2: partition2, delay: delay}
	return &network
}

func (n *PartioningNetwork) Connect(adress string) {}
func (n *PartioningNetwork) RegisterClient(client *ClientInterface) {
	n.clients = append(n.clients, client)
}

func (n *PartioningNetwork) Deliver(data interface{}, client *ClientInterface) {
	receiver := signature.PublicKeyToString((*client).GetIdentity())
	randDelay := rand.Float32() * 2
	switch data := data.(type) {
	case *types.Block:
		if _, exists := n.partition1[data.PublicKey]; exists {
			if _, exists := n.partition2[receiver]; exists {
				// delay, across partions (p1 -> p2)
				time.Sleep(time.Duration(float32(n.delay)+(randDelay)) * time.Second)
			} // no delay (p1 -> p1)
		} else {
			if _, exists := n.partition1[receiver]; exists {
				// delay, across partions (p2 -> p1)
				time.Sleep(time.Duration(float32(n.delay)+(randDelay)) * time.Second)
			} // no delay (p2 -> p2)
		}
	case types.Message:
		if _, exists := n.partition1[data.Peer]; exists {
			if _, exists := n.partition2[receiver]; exists {
				// delay, across partions (p1 -> p2)
				time.Sleep(time.Duration(float32(n.delay)+(randDelay)) * time.Second)
			} // no delay (p1 -> p1)
		} else {
			if _, exists := n.partition1[receiver]; exists {
				// delay, across partions (p2 -> p1)
				time.Sleep(time.Duration(float32(n.delay)+(randDelay)) * time.Second)
			} // no delay (p2 -> p2)
		}
		// data.
		// 	(*c.finalizer).HandleMessage(&data)
	}
	(*client).Output(data)
}

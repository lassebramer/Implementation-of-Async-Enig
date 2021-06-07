package network

import (
	"math/rand"
	"time"
)

type DelayNetwork struct {
	clients []*ClientInterface
	delay   int
}

func MakeDelayNetwork(delay int) Network {
	network := DelayNetwork{clients: make([]*ClientInterface, 0), delay: delay}
	return &network
}

func (n *DelayNetwork) Send(data interface{}) {
	for _, client := range n.clients {
		go n.Deliver(data, client)
	}
}
func (n *DelayNetwork) Connect(adress string) {}
func (n *DelayNetwork) RegisterClient(client *ClientInterface) {
	n.clients = append(n.clients, client)
}

func (n *DelayNetwork) Deliver(data interface{}, client *ClientInterface) {
	// Delay here man
	coinflip := rand.Intn(2) == 1
	for coinflip {
		time.Sleep(time.Duration(n.delay) * time.Millisecond)
		coinflip = rand.Intn(2) == 1
	}
	(*client).Output(data)
}

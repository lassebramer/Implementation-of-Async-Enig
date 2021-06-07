package network

import (
	"math/rand"
)

type DropNetwork struct {
	clients    []*ClientInterface
	dropChance int
}

func MakeDropNetwork(dropChance int) Network {
	network := DropNetwork{clients: make([]*ClientInterface, 0), dropChance: dropChance}
	return &network
}

func (n *DropNetwork) Send(data interface{}) {
	for _, client := range n.clients {
		go n.Deliver(data, client)
	}
}
func (n *DropNetwork) Connect(adress string) {}
func (n *DropNetwork) RegisterClient(client *ClientInterface) {
	n.clients = append(n.clients, client)
}

func (n *DropNetwork) Deliver(data interface{}, client *ClientInterface) {
	drop := rand.Intn(100) < n.dropChance
	if !drop {
		(*client).Output(data)
	}

}

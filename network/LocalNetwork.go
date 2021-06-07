package network

type LocalNetwork struct {
	clients []*ClientInterface
}

func (n *LocalNetwork) Send(data interface{}) {
	for _, client := range n.clients {
		go n.Deliver(data, client)
	}
}
func (n *LocalNetwork) Connect(adress string) {}
func (n *LocalNetwork) RegisterClient(client *ClientInterface) {
	n.clients = append(n.clients, client)
}

func (n *LocalNetwork) Deliver(data interface{}, client *ClientInterface) {
	(*client).Output(data)
}

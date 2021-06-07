package network

import "crypto/rsa"

type Network interface {
	Send(data interface{})
	Connect(adress string)
	RegisterClient(client *ClientInterface)
}

type ClientInterface interface {
	Output(data interface{})
	Send(data interface{})
	GetIdentity() rsa.PublicKey
}

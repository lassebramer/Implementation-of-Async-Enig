package blockchain

import (
	"crypto/rsa"

	"github.com/Hvassaa/bachelor/network"
	"github.com/Hvassaa/bachelor/types"
)

type DummyStruct struct {
	message string
}

type Client struct {
	network    *network.Network
	blockchain *Blockchain
	finalizer  *Finalizer
	identity   rsa.PublicKey
}

func MakeClient(network *network.Network, blockchain *Blockchain, finalizer *Finalizer, key rsa.PublicKey) network.ClientInterface {
	client := Client{network: network, blockchain: blockchain, finalizer: finalizer, identity: key}
	return &client
}

func (c *Client) Output(data interface{}) {
	switch data := data.(type) {
	case types.SignedTransaction:
		if c.blockchain == nil {
			return
		}
		(*c.blockchain).HandleTransaction(data)
	case *types.Block:
		(*c.blockchain).HandleBlock(data)
	case types.Message:
		(*c.finalizer).HandleMessage(&data)
	}
}

func (c *Client) Send(data interface{}) {
	(*c.network).Send(data)
}

func (c *Client) GetIdentity() rsa.PublicKey {
	return c.identity
}

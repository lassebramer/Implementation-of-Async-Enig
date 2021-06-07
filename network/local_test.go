package network

// import (
// 	"testing"

// )

// func TestDummyChain(t *testing.T) {
// 	net := LocalNetwork{}
// 	client := &Client{network: &net}
// 	net.RegisterClient(client)
// 	dummyChain := &blockchain.DummyChain{}
// 	client.blockchain = dummyChain
// 	dummystruct := blockchain.SignedTransaction{}
// 	net.Send(dummystruct)
// 	net.Deliver(0, 0)
// 	if len(dummyChain.Transactions) == 0 {
// 		t.Errorf("Message did not arrive")
// 	}
// }

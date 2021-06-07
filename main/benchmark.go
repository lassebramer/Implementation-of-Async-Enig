package main

import (
	"crypto/rsa"
	"time"

	"github.com/Hvassaa/bachelor/blockchain"
	"github.com/Hvassaa/bachelor/network"
	"github.com/Hvassaa/bachelor/signature"
)

func main() {
	Benchmark(10)
}

func Benchmark(noOfPeers int) {
	var LocalNetwork network.Network = &network.LocalNetwork{}

	//hardnees is a dummy-value
	hardness := 500
	privateKeys := []rsa.PrivateKey{}
	publicKeys := []rsa.PublicKey{}
	protocols := []*blockchain.Blockchain{}

	//create keypairs
	for i := 0; i < noOfPeers; i++ {
		privateKey := signature.KeyGen(4096)
		publicKey := privateKey.PublicKey
		publicKeys = append(publicKeys, publicKey)
		privateKeys = append(privateKeys, *privateKey)
	}

	for _, privateKey := range privateKeys {
		protocol := blockchain.MakeProtocol(privateKey, publicKeys, hardness)
		protocols = append(protocols, &protocol)
		protocol.ReadTree("out/writtenTree.gob")
		finalizer := blockchain.MakeFinalizer(&protocol, privateKey)
		protocol.SetFinalizer(finalizer)
		//TODO check pointers på network, protocol og finalizer
		var Client network.ClientInterface = blockchain.MakeClient(&LocalNetwork, &protocol, finalizer)
		LocalNetwork.RegisterClient(&Client)
		protocol.RegisterClient(&Client)
		finalizer.RegisterClient(&Client)
	}

	for _, proto := range protocols {
		f := (*proto).GetFinalizer()
		f.StartFinalizer()
	}
	time.Sleep(60 * time.Second)

	td := blockchain.TreeDrawer{Bc: protocols[0], ID: "bench"}
	td.DrawTree()

}

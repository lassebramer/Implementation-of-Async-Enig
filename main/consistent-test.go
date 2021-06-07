package main

import (
	"crypto/rsa"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Hvassaa/bachelor/blockchain"
	"github.com/Hvassaa/bachelor/network"
	"github.com/Hvassaa/bachelor/signature"
)

func main() {
	args := os.Args
	duration, err := strconv.Atoi(args[1])
	checkErr(err)
	noOfPeers, err := strconv.Atoi(args[2])
	checkErr(err)
	hardness, err := strconv.Atoi(args[3])
	checkErr(err)
	finalizeInt, err := strconv.Atoi(args[4])
	finalize := finalizeInt == 1
	checkErr(err)
	networkType, err := strconv.Atoi(args[5])
	checkErr(err)
	networkParam, err := strconv.Atoi(args[6])
	checkErr(err)
	testRun(duration, noOfPeers, hardness, finalize, networkType, networkParam)
	// testTree()
}

func checkErr(err error) {
	if err != nil {
		fmt.Println("Got error ", err)
		fmt.Println("Give arguments as integer: duration (secs), no. of peers, hardness, whether to use finalizer (1 for yes), networkType 1:normal 2:delay 3:drop 4:partition, network params")
		os.Exit(1)
	}
}

func testRun(duration, noOfPeers, hardness int, finalize bool, networkType int, networkParam int) {

	// privateKeys := []rsa.PrivateKey{}
	publicKeys := []rsa.PublicKey{}
	protocols := []*blockchain.Blockchain{}

	// read / create keypairs
	privateKeys := signature.ReadKeys("keys/keys.gob", noOfPeers)
	for _, key := range privateKeys {
		publicKey := key.PublicKey
		publicKeys = append(publicKeys, publicKey)
	}
	var Network network.Network
	var networkString string
	switch networkType {
	case 1:
		Network = &network.LocalNetwork{}
		networkString = "Normal"
	case 2:
		Network = network.MakeDelayNetwork(networkParam)
		networkString = "Delay"
	case 3:
		Network = network.MakeDropNetwork(networkParam)
		networkString = "Drop"
	case 4:
		Network = network.MakePartioningNetwork(publicKeys, networkParam)
		networkString = "Partition"
	}

	// Network := network.MakePartioningNetwork(publicKeys)

	//create protocols and clients and register them
	for _, privateKey := range privateKeys {
		protocol := blockchain.MakeProtocol(privateKey, publicKeys, hardness)
		protocols = append(protocols, &protocol)
		finalizer := blockchain.MakeFinalizer(&protocol, privateKey)
		if finalize {
			protocol.SetFinalizer(finalizer)
		}
		//TODO check pointers på network, protocol og finalizer
		var Client network.ClientInterface = blockchain.MakeClient(&Network, &protocol, finalizer, privateKey.PublicKey)
		Network.RegisterClient(&Client)
		protocol.RegisterClient(&Client)
		if finalize {
			finalizer.RegisterClient(&Client)
		}
	}
	// time.Sleep(5 * time.Second)

	for _, proto := range protocols {
		(*proto).StartLottery()
	}

	// noOfProtos := len(protocols)
	// go func() {
	// 	for {
	// 		from := rand.Intn(noOfProtos)
	// 		to := rand.Intn(noOfProtos)
	// 		if from == to {
	// 			continue
	// 		}
	// 		randAmountToSend := rand.Intn(200)
	// 		// randAmountToSend := 5
	// 		(*protocols[from]).MakeTransaction(publicKeys[to], randAmountToSend)
	// 		// fmt.Printf("to: %d, from: %d, amount: %d\n", to+1, from+1, randAmountToSend)
	// 		time.Sleep(200 * time.Millisecond)
	// 	}
	// }()

	if finalize {
		for _, proto := range protocols {
			f := (*proto).GetFinalizer()
			f.StartFinalizer()
		}
	}
	round := 1
	// adjust rounds
	for i := 0; i < 1; i++ {
		time.Sleep(time.Duration(duration) * time.Second)
		netParam := strconv.Itoa(networkParam)

		// draw tree
		if finalize {
			for i := 0; i < 5; i++ {
				td := blockchain.TreeDrawer{Bc: protocols[i], ID: "Final-" + networkString + netParam + strconv.Itoa(i), CopyOfTree: (*(*protocols[i]).GetBlockTree()).GetCopyOfTree()}
				td.DrawTree()
			}

			// td := blockchain.TreeDrawer{Bc: protocols[0], ID: "Final-" + networkString + netParam, CopyOfTree: (*(*protocols[0]).GetBlockTree()).GetCopyOfTree()}
			// // td.DrawTree()
			// fname := "out/" + strconv.Itoa(noOfPeers) + "Finalizer-" + networkString + ".csv"
			// td.WriteLengths(fname, []string{netParam})
			// fname = "out/" + strconv.Itoa(noOfPeers) + "DistancesToLastFinal-" + networkString + ".csv"
			// td.WriteDistancesToLastFinals(fname, []string{netParam})
			// FinAndStbDist := (*protocols[0]).GetFinalizer().FinalAndSTBDistances
			// fname = "out/" + strconv.Itoa(noOfPeers) + "FinalAndSTBDistances-" + networkString + ".csv"
			// FinAndStbDist = append([]string{netParam}, FinAndStbDist...)
			// blockchain.WriteSliceToCSV(fname, FinAndStbDist)
		} else {
			td := blockchain.TreeDrawer{Bc: protocols[0], ID: "NotFinal" + networkString + netParam, CopyOfTree: (*(*protocols[0]).GetBlockTree()).GetCopyOfTree()}
			td.DrawTree()
			fname := "out/" + strconv.Itoa(noOfPeers) + "-" + networkString + ".csv"
			td.WriteLengths(fname, []string{netParam})
		}
		round++
	}
}

//test read write
func testTree() {
	privateKey := signature.KeyGen(4096)
	publicKey := []rsa.PublicKey{privateKey.PublicKey}

	protocol := blockchain.MakeProtocol(*privateKey, publicKey, 950)
	protocol.ReadTree("out/writtenTree.gob")
	td := blockchain.TreeDrawer{Bc: &protocol, ID: strconv.Itoa(10)}
	td.DrawTree()
}

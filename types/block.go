package types

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"strconv"

	"github.com/Hvassaa/bachelor/signature"
)

type (
	Block struct {
		Slot                int
		SlotSignature       []byte
		Transactions        []SignedTransaction
		PublicKey           string
		Signature           []byte
		Parent              Hash
		Height              int
		DistanceToLastFinal int
	}
)

func (block *Block) Hash() Hash {
	bytes := block.toBytes()
	hash := sha256.Sum256(bytes)
	return hash
}

func (block *Block) Sign(sk *rsa.PrivateKey) {
	bytes := block.toBytes()
	signature := signature.Sign(sk, bytes)
	block.Signature = signature
}

func (b *Block) toBytes() []byte {
	// fmt.Println(b)
	s := []byte(strconv.Itoa(b.Slot) + ";" + b.PublicKey)
	s = append(s, b.Parent[:]...)
	for _, t := range b.Transactions {
		s = append(s, []byte(t.ID)...)
	}
	//we do not use time as it is determined by slot
	//we do not use signature as it will be determined by the other fields
	return s
}

func (block *Block) VerifyBlock() bool {
	blockKey := signature.PublicKeyFromString(block.PublicKey)
	hash := block.Hash()

	err := rsa.VerifyPSS(&blockKey, crypto.SHA256, hash[:], block.Signature, nil)

	return err == nil

	// if err != nil {
	// 	fmt.Println("Blocksignature", err)
	// 	return false
	// }
	// return true
}

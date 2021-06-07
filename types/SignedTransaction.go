package types

import (
	"crypto/rsa"
	"strconv"

	"github.com/Hvassaa/bachelor/signature"
)

type SignedTransaction struct {
	ID        string // Any string
	From      string // A verification key coded as a string
	To        string // A verification key coded as a string
	Amount    int    // Amount to transfer
	Signature []byte // Potential signature coded as string
}

func NewSignedTransaction(id string, fromPrivKey rsa.PrivateKey, to string, amount int) SignedTransaction {
	publicKeyString := signature.PublicKeyToString(fromPrivKey.PublicKey)
	transactionString := id + publicKeyString + to + strconv.Itoa(amount)
	sig := signature.Sign(&fromPrivKey, []byte(transactionString))
	return SignedTransaction{
		ID:        id,
		From:      publicKeyString,
		To:        to,
		Amount:    amount,
		Signature: sig,
	}
}

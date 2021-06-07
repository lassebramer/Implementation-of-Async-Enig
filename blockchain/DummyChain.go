package blockchain

import "github.com/Hvassaa/bachelor/types"

type DummyChain struct {
	Transactions []types.SignedTransaction
}

func (d *DummyChain) HandleTransaction(transaction types.SignedTransaction) {
	d.Transactions = append(d.Transactions, transaction)
}

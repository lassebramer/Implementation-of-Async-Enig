package blockchain

import (
	"crypto/rsa"
	"strconv"

	"github.com/Hvassaa/bachelor/signature"
	"github.com/Hvassaa/bachelor/types"
)

type Ledger struct {
	Accounts map[string]int

	// lock     sync.Mutex
	Total int
}

func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

func (l *Ledger) Reward(to string, amount int) {
	// l.lock.Lock()

	l.Total += amount
	l.Accounts[to] += amount

	// l.lock.Unlock()
}

func (l *Ledger) SignedTransaction(t *types.SignedTransaction) bool {
	// l.lock.Lock()
	// defer l.lock.Unlock()
	/* We verify that the t.Signature is a valid RSA
	* signature on the rest of the fields in t under
	* the public key t.From.
	 */

	transactionString := t.ID + t.From + t.To + strconv.Itoa(t.Amount)
	transactionBytes := []byte(transactionString)
	key := signature.PublicKeyFromString(t.From)

	validSignature := signature.Verify(&key, transactionBytes, t.Signature)

	if validSignature && t.Amount > 0 && (l.Accounts[t.From]-t.Amount) >= 0 {
		l.Accounts[t.From] -= t.Amount
		l.Accounts[t.To] += t.Amount
		return true
	}
	return false
}

func (l *Ledger) Copy() *Ledger {
	res := MakeLedger()
	res.Total = l.Total
	for acc, amount := range l.Accounts {
		res.Accounts[acc] = amount
	}
	return res
}

func (l *Ledger) ForceTransaction(from string, to string, amount int) {
	// l.lock.Lock()

	l.Accounts[from] -= amount

	l.Accounts[to] += amount

	// l.lock.Unlock()
}

func (l *Ledger) ShowLedger(accounts []rsa.PublicKey) {
	accountId := 1
	for _, pk := range accounts {
		publicString := signature.PublicKeyToString(pk)
		println(accountId, " | ", l.Accounts[publicString])
		accountId++
	}
}

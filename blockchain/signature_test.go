package blockchain

import (
	"testing"

	"github.com/Hvassaa/bachelor/signature"
	"github.com/Hvassaa/bachelor/types"
)

func TestSignature(t *testing.T) {
	privateKey := signature.KeyGen(4096)
	message := []byte("12345677899545")
	sig := signature.Sign(privateKey, message)
	verified := signature.Verify(&privateKey.PublicKey, message, sig)
	if !verified {
		t.Errorf("Valid signature and key was rejected")
	}
	privateKey2 := signature.KeyGen(4096)
	badVerify := signature.Verify(&privateKey2.PublicKey, message, sig)
	if badVerify {
		t.Errorf("Bad signature/key combo passed")
	}
}

func TestLedger(t *testing.T) {
	privateKey1 := signature.KeyGen(4096)
	privateKey2 := signature.KeyGen(4096)
	publicKey1 := signature.PublicKeyToString(privateKey1.PublicKey)
	publicKey2 := signature.PublicKeyToString(privateKey2.PublicKey)

	ledger := MakeLedger()
	ledger.Accounts[publicKey1] = 10
	ledger.Accounts[publicKey2] = 0
	trans1 := types.NewSignedTransaction("et godt id", *privateKey1, publicKey2, 10)
	ledger.SignedTransaction(&trans1)
	acc1updated := ledger.Accounts[publicKey1] == 0
	acc2updated := ledger.Accounts[publicKey2] == 10
	if !acc1updated {
		t.Errorf("Account '1' for %s not updated", publicKey1)
	}
	if !acc2updated {
		t.Errorf("Account '2' for %s not updated", publicKey2)
	}
}

// func testTree(t *testing.T) {
// 	proto = MakeProtocol()
// }

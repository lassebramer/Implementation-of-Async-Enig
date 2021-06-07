package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
)

// KEYS should be ~4096 bits

func PublicKeyToString(key rsa.PublicKey) string {
	return key.N.String() + ";" + strconv.Itoa(key.E)
}

func PublicKeyFromString(str string) rsa.PublicKey {
	bigStrings := strings.Split(str, ";")
	n := new(big.Int)
	n, ok := n.SetString(bigStrings[0], 10)
	if !(ok) {
		fmt.Println("Error converting string to bigInt")
	}
	e, ok2 := strconv.Atoi(bigStrings[1])
	if ok2 != nil {
		fmt.Println(ok2)
	}

	return rsa.PublicKey{N: n, E: e}
}

func KeyGen(k int) *rsa.PrivateKey {
	key, _ := rsa.GenerateKey(rand.Reader, k)
	key.Precompute()
	return key
}

func Sign(key *rsa.PrivateKey, message []byte) []byte {
	hashedMessageBytes := sha256.Sum256(message)
	signature, err := rsa.SignPSS(rand.Reader, key, crypto.SHA256, hashedMessageBytes[0:], nil)
	if err != nil {
		fmt.Println(err)
	}
	return signature
}

func Verify(key *rsa.PublicKey, message []byte, signedMessage []byte) bool {
	hashedMessageBytes := sha256.Sum256(message)

	err := rsa.VerifyPSS(key, crypto.SHA256, hashedMessageBytes[0:], signedMessage, nil)

	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func WriteKeys(numOfKeys int) {
	os.Mkdir("out", os.ModePerm)
	os.Mkdir("out/keys", os.ModePerm)
	file, err := os.Create("out/keys/keys.gob")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	keys := []rsa.PrivateKey{}
	for i := 0; i < numOfKeys; i++ {
		privateKey := KeyGen(4096)
		keys = append(keys, *privateKey)
	}

	encoder.Encode(keys)
}

func ReadKeys(filename string, numOfKeys int) []rsa.PrivateKey {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	keys := []rsa.PrivateKey{}

	decoder := gob.NewDecoder(file)
	decoder.Decode(&keys)
	return keys[0:numOfKeys]
}

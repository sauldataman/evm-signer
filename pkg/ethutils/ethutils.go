package ethutils

import (
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

// Account represents an Ethereum account
type Account struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

// GetAccountFromMnemonic derives an account from a mnemonic at the given index
func GetAccountFromMnemonic(mnemonic string, index int) *Account {
	seed := bip39.NewSeed(mnemonic, "")

	// Derive the master key
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil
	}

	// Standard Ethereum derivation path: m/44'/60'/0'/0/index
	path := accounts.DefaultBaseDerivationPath
	path[len(path)-1] = uint32(index)

	key := masterKey
	for _, n := range path {
		key, err = key.Child(n)
		if err != nil {
			return nil
		}
	}

	privateKey, err := key.ECPrivKey()
	if err != nil {
		return nil
	}

	privateKeyECDSA := privateKey.ToECDSA()
	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Account{
		Address:    address,
		PrivateKey: privateKeyECDSA,
	}
}

// GetAccountFromPStr creates an account from a private key string
func GetAccountFromPStr(privateKeyStr string) *Account {
	// Remove 0x prefix if present
	if len(privateKeyStr) > 2 && privateKeyStr[:2] == "0x" {
		privateKeyStr = privateKeyStr[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Account{
		Address:    address,
		PrivateKey: privateKey,
	}
}

// Sign signs a message with the given private key
func Sign(message []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	hash := accounts.TextHash(message)
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, err
	}

	// Adjust V value for Ethereum compatibility (add 27)
	if signature[64] < 27 {
		signature[64] += 27
	}

	return signature, nil
}

package account

import (
	"cs-evm-signer/types"
	"fmt"
	_keystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"os"
)

type keystore struct {
	path   string
	pass   string
	priKey string
}

func NewKeystore(path, pass string) (*keystore, error) {
	if path == "" {
		return nil, fmt.Errorf("keystore config error")
	}
	return &keystore{path: path, pass: pass}, nil
}

func (k *keystore) Decrypt() []*types.Account {
	keyJson, err := os.ReadFile(k.path)
	if err != nil {
		logger.Fatalf("Failed to read the keyfile at '%s': %v", k.path, err)
	}

	key, err := _keystore.DecryptKey(keyJson, k.pass)
	if err != nil {
		utils.Fatalf("Error decrypting key: %v", err)
	}

	logger.Infof("address: [%s]", key.Address.String())
	var accounts []*types.Account
	account := &types.Account{
		Address: key.Address,
		PriKey:  key.PrivateKey,
	}
	accounts = append(accounts, account)
	return accounts
}

func (k *keystore) Crypto() error {
	return fmt.Errorf("unSupport crypto")
}

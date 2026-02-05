package account

import (
	"cs-evm-signer/types"
	"encoding/json"
	"fmt"
	ethutils "github.com/CoinSummer/go-ethutils"
	_keystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/google/uuid"
	"os"
	"strconv"

	dString "github.com/CoinSummer/go-utils/go/string"
)

type encryptedMnemonic struct {
	mnemonic string
	password string
	indexMap map[int64]struct{}
}

type encryptedKeyJSONV3 struct {
	Address string               `json:"address"`
	Crypto  _keystore.CryptoJSON `json:"crypto"`
	Id      string               `json:"id"`
	Version int                  `json:"version"`
}

const version = 3

func NewEncryptedMnemonic(key, password, indexRange string) (*encryptedMnemonic, error) {
	if key == "" || password == "" || indexRange == "" {
		return nil, fmt.Errorf("encrypted mnemonic config error")
	}

	indexMap := make(map[int64]struct{})
	splitMap, err := dString.SplitNum(indexRange)
	if err != nil {
		return nil, fmt.Errorf("index config error")
	}

	for _index := range splitMap {
		__index, _ := strconv.ParseInt(_index, 10, 64)
		indexMap[__index] = struct{}{}
	}
	return &encryptedMnemonic{mnemonic: key, password: password, indexMap: indexMap}, nil
}

func (m *encryptedMnemonic) Decrypt() []*types.Account {
	var accounts []*types.Account

	keyJson, err := os.ReadFile(m.mnemonic)
	if err != nil {
		logger.Fatalf("Failed to read the keyfile at '%s': %v", m.mnemonic, err)
	}

	key := new(encryptedKeyJSONV3)
	if err = json.Unmarshal(keyJson, key); err != nil {
		logger.Fatalf("Unmarshal  keyfile at '%s': %v", m.mnemonic, err)
	}

	keyBytes, _, err := decryptKeyV3(key, m.password)
	if err != nil {
		utils.Fatalf("Error decrypting key: %v", err)
	}

	for k := range m.indexMap {
		account := &types.Account{
			Index: k,
		}

		_key := ethutils.GetAccountFromMnemonic(string(keyBytes), int(k))
		account.Address = _key.Address
		logger.Debugf("index: [%d], address: [%s] \n", k, account.Address.String())
		account.PriKey = _key.PrivateKey

		accounts = append(accounts, account)
	}
	return accounts
}

func decryptKeyV3(keyProtected *encryptedKeyJSONV3, auth string) (keyBytes []byte, keyId []byte, err error) {
	if keyProtected.Version != version {
		return nil, nil, fmt.Errorf("version not supported: %v", keyProtected.Version)
	}
	keyUUID, err := uuid.Parse(keyProtected.Id)
	if err != nil {
		return nil, nil, err
	}
	keyId = keyUUID[:]
	plainText, err := _keystore.DecryptDataV3(keyProtected.Crypto, auth)
	if err != nil {
		return nil, nil, err
	}
	return plainText, keyId, err
}

func (m *encryptedMnemonic) Crypto() error {
	return fmt.Errorf("unSupport crypto")
}

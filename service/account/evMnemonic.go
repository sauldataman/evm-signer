package account

import (
	"crypto/ecdsa"
	"cs-evm-signer/pkg/hsm"
	"cs-evm-signer/types"
	"fmt"
	ethutils "github.com/CoinSummer/go-ethutils"
	_keystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"sort"
)

type evMnemonic struct {
	keys map[int64]interface{}
}

func NewEvMnemonic(keys map[int64]interface{}) (*evMnemonic, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("evMnemonic config error")
	}
	return &evMnemonic{keys: keys}, nil
}

func (em *evMnemonic) Decrypt() []*types.Account {
	var accounts []*types.Account
	lastPass := ""

	keys := em.sortKeys()
	for index, k := range keys {
		subKeyMap := em.keys[int64(k)].(map[string]interface{})
		account := &types.Account{
			Index: int64(k),
		}

		var _address common.Address
		var _priKey *ecdsa.PrivateKey
		// check type
		subKeyType := subKeyMap["type"].(string)
		switch _AccountType(subKeyType) {
		case PlainPrivateKeyTy:
			err := checkKeystoreParams(subKeyMap)
			if err != nil {
				logger.Fatalf("subKey config for acount index %d error: %s", k, err)
			}

			_account := ethutils.GetAccountFromPStr(subKeyMap["key"].(string))
			if _account == nil {
				logger.Errorf("plain privateKey is invalidate")
				return nil
			}
			_address = _account.Address
			_priKey = _account.PrivateKey
		case KeyStoreTy:
			err := checkKeystoreParams(subKeyMap)
			if err != nil {
				logger.Fatalf("subKey config for account index %d error: %s", k, err)
			}

			_key := subKeyMap["key"].(string)
			keyJson, err := os.ReadFile(_key)
			if err != nil {
				logger.Fatalf("Failed to read the keyfile at '%s': %v", _key, err)
			}

			pass := subKeyMap["pass"].(string)
			lastPass, pass = resetPass(lastPass, pass, index, subKeyMap)
			__keystore, err := _keystore.DecryptKey(keyJson, pass)
			if err != nil {
				utils.Fatalf("Error decrypting key: %v", err)
			}
			_address = __keystore.Address
			_priKey = __keystore.PrivateKey
		case PlainMnemonicTy:
			err := checkMnemonicParams(subKeyMap)
			if err != nil {
				logger.Fatalf("subKey config for account index %d error: %s", k, err)
			}
			mnemonic := subKeyMap["key"].(string)
			mnemonicIndex := subKeyMap["index"].(string)
			// PlainMnemonic in EvMnemonic uses the first index from the range
			pm, err := NewPlainMnemonic(mnemonic, mnemonicIndex)
			if err != nil {
				logger.Fatalf("create plain mnemonic error: %s", err)
			}
			pmAccounts := pm.Decrypt()
			if len(pmAccounts) == 0 {
				logger.Fatalf("plain mnemonic decrypt failed for account index %d", k)
			}
			// Use the first account from the mnemonic
			_address = pmAccounts[0].Address
			_priKey = pmAccounts[0].PriKey
		case HSMTy:
			err := checkHsmParams(subKeyMap)
			if err != nil {
				logger.Fatalf("subKey config for account index %d error: %s", k, err)
			}
			_hsmParams := unWrapHsmData(subKeyMap)
			lastPass, _hsmParams.pin = resetPass(lastPass, _hsmParams.pin, index, subKeyMap)

			hsmParam := &hsm.ParamsHsm{
				Url:          _hsmParams.url,
				Provider:     hsm.ProviderHsm(_hsmParams.provider),
				LibPath:      "",
				Pin:          _hsmParams.pin,
				PublicKeyId:  _hsmParams.publicKeyId,
				PrivateKeyId: _hsmParams.privateKeyId,
			}

			fmt.Printf("_hsmParams.url: %s \n", _hsmParams.url)
			hc, err := NewHsm(hsmParam)
			if err != nil {
				logger.Fatalf("create hsm client error: %s", err)
			}
			_, address, err := hc.client.GetPublicKey()
			if err != nil {
				logger.Fatalf("get address via HSM error: %s", err)
			}
			_address = address
			account.HsmObjId = int64(_hsmParams.privateKeyId)
			account.HsmClient = hc.client
		default:
			logger.Fatalf("%s type unsupported", subKeyType)
		}

		logger.Infof("account type: [%s], index: [%d], address: [%s]", subKeyType, k, _address)
		account.Address = _address
		account.PriKey = _priKey
		accounts = append(accounts, account)
	}
	return accounts
}

func (em *evMnemonic) sortKeys() []int {
	var keys []int
	for k := range em.keys {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	return keys
}

func resetPass(lastPass, pass string, k int, subKeyMap map[string]interface{}) (string, string) {
	if _, ok := subKeyMap["use_last_pass"]; !ok {
		subKeyMap["use_last_pass"] = false
	}

	isUseLastPass := subKeyMap["use_last_pass"].(bool)
	if isUseLastPass {
		if k == 0 {
			pass = passPhrase(KeyStoreTy, pass)
			lastPass = pass
		} else if k != 0 && pass == "" {
			pass = lastPass
		}
	} else {
		pass = passPhrase(KeyStoreTy, pass)
		lastPass = pass
	}
	return lastPass, pass
}

func (em *evMnemonic) Crypto() error {
	return fmt.Errorf("unSupport crypto")
}

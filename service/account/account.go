package account

import (
	"crypto/ecdsa"
	"evm-signer/pkg/ethutils"
	"evm-signer/pkg/logging"
	"evm-signer/types"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-errors/errors"
	"reflect"
	"strconv"
	"strings"
)

type _AccountType string

const (
	KeyStoreTy          _AccountType = "Keystore"
	EvMnemonicTy        _AccountType = "EvMnemonic"
	EncryptedMnemonicTy _AccountType = "EncryptedMnemonic"
	PlainMnemonicTy     _AccountType = "PlainMnemonic"
	PlainPrivateKeyTy   _AccountType = "PlainPrivateKey"
)

var (
	errKeyField   = errors.New("account config error, must contains key field")
	errIndexField = errors.New("account config error, must contains index field")

	errKeyArgument   = errors.New("account config error, key must be string")
	errPassArgument  = errors.New("account config error, pass must be string")
	errUseLastPass   = errors.New("account config error, use_last_pass must be bool")
	errIndexArgument = errors.New("account config error, index must be string, eg: 0-9,256")

	errKeyNull   = errors.New("key field is null")
	errIndexNull = errors.New("index field is null")
)

var logger *logging.SugaredLogger

func SetLogger(_logger *logging.SugaredLogger) {
	logger = _logger
}

type Account struct {
	accountTy _AccountType
	priKey    *ecdsa.PrivateKey
	params    map[string]interface{}
}

type _AccountOpt struct {
	accountTy _AccountType
	params    interface{}
}

type _AccountInfo struct{}

func (a _AccountInfo) Index(accountMap map[int64]*types.Account, index int64) (*types.Account, bool) {
	account, ok := accountMap[index]
	if !ok {
		return nil, ok
	}

	if account.PriKey == nil {
		return nil, false
	}
	return account, true
}

func (a _AccountInfo) Address(accountMap map[string]*types.Account, address string) (*types.Account, bool) {
	account, ok := accountMap[strings.ToLower(address)]
	if !ok {
		return nil, ok
	}
	if account.Address.Hex() == "" {
		return nil, false
	}
	if account.PriKey == nil {
		return nil, false
	}
	return account, true
}

func (a *_AccountOpt) GetAccount() IGetAccount {
	return &_AccountInfo{}
}

type Crypto struct {
	accountTy _AccountType
	params    interface{}
}

// key, pass
func checkKeystoreParams(params map[string]interface{}) error {
	if _, ok := params["key"]; !ok {
		return errKeyField
	}

	// key 不应该为空
	if params["key"] == nil {
		return errKeyNull
	}

	if reflect.TypeOf(params["key"]).Kind() != reflect.String {
		return errKeyArgument
	}

	// pass 为选填字段：用户没有在配置文件中填写 pass 时的处理。
	if _, ok := params["pass"]; !ok {
		params["pass"] = ""
	}

	// 用户填写的 pass 字段为空时的处理。
	if params["pass"] == nil {
		params["pass"] = ""
	}

	if reflect.TypeOf(params["pass"]).Kind() != reflect.String {
		return errPassArgument
	}

	// use_last_pass 为选填字段：用户没有在配置文件中填写 use_last_pass 时的处理。
	if _, ok := params["use_last_pass"]; !ok {
		params["use_last_pass"] = false
	}

	// 用户填写的 use_last_pass 字段为空时的处理。
	if params["use_last_pass"] == nil {
		params["use_last_pass"] = false
	}

	if reflect.TypeOf(params["use_last_pass"]).Kind() != reflect.Bool {
		return errUseLastPass
	}
	return nil
}

// key, index
func checkMnemonicParams(params map[string]interface{}) error {
	if _, ok := params["key"]; !ok {
		return errKeyField
	}

	// key 不应该为空
	if params["key"] == nil {
		return errKeyNull
	}

	if reflect.TypeOf(params["key"]).Kind() != reflect.String {
		return errKeyArgument
	}

	if _, ok := params["index"]; !ok {
		return errIndexField
	}

	// index 不应该为空
	if params["index"] == nil {
		return errIndexNull
	}

	if reflect.TypeOf(params["index"]).Kind() != reflect.String {
		return errIndexArgument
	}

	// pass 为选填字段：用户没有在配置文件中填写 pass 时的处理。
	if _, ok := params["pass"]; !ok {
		params["pass"] = ""
	}

	// 用户填写的 pass 字段为空时的处理。
	if params["pass"] == nil {
		params["pass"] = ""
	}

	if reflect.TypeOf(params["pass"]).Kind() != reflect.String {
		return errPassArgument
	}
	return nil
}

func (a *_AccountOpt) Crypto() (ICrypto, error) {
	c := &Crypto{
		accountTy: a.accountTy,
		params:    a.params,
	}

	switch c.accountTy {
	case KeyStoreTy:
		params := c.params.(map[string]interface{})

		err := checkKeystoreParams(params)
		if err != nil {
			return nil, err
		}
		key := params["key"].(string)
		pass := passPhrase(KeyStoreTy, params["pass"].(string))
		return NewKeystore(key, pass)
	case EvMnemonicTy:
		params := c.params.(map[string]interface{})
		if _, ok := params["keys"]; !ok {
			return nil, fmt.Errorf("account config error, EvMnemonic Type must contains keys field")
		}

		key := make(map[int64]interface{})
		for paramType, param := range params {
			if paramType != "keys" {
				continue
			}

			for index, val := range param.(map[string]interface{}) {
				i64Index, err := strconv.ParseInt(index, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("EvMnemonic keys index must be number")
				}
				key[i64Index] = val
			}
		}
		return NewEvMnemonic(key)
	case EncryptedMnemonicTy:
		params := c.params.(map[string]interface{})
		err := checkMnemonicParams(params)
		if err != nil {
			return nil, err
		}
		key := params["key"].(string)
		indexRange := params["index"].(string)
		pass := passPhrase(EncryptedMnemonicTy, params["pass"].(string))
		return NewEncryptedMnemonic(key, pass, indexRange)
	case PlainMnemonicTy:
		params := c.params.(map[string]interface{})
		err := checkMnemonicParams(params)
		if err != nil {
			return nil, err
		}
		key := params["key"].(string)
		index := params["index"].(string)
		return NewPlainMnemonic(key, index)
	case PlainPrivateKeyTy:
		params := c.params.(map[string]interface{})
		if _, ok := params["key"]; !ok {
			return nil, fmt.Errorf("account config error, PlainPrivateKey Type must contains key field")
		}
		key := params["key"].(string)
		return NewPlainPrivateKey(key)
	default:
		return nil, fmt.Errorf("unSupported account type, only support Keystore, EvMnemonic, " +
			"EncryptedMnemonic, PlainMnemonic, PlainPrivateKey at the moment")
	}
}

func passPhrase(accountTy _AccountType, pass string) string {
	if pass == "" {
		password := getPassPhrase(fmt.Sprintf(
			"please enter password for [%s] account",
			accountTy,
		), false)
		if password == "" {
			logger.Fatalf("password does not empty")
		}
		pass = password
	}
	return pass
}

func (a *Account) Account() IAccountOpt {
	return &_AccountOpt{
		accountTy: a.accountTy,
		params:    a.params,
	}
}

func (a *Account) GetPriKey() *ecdsa.PrivateKey {
	return a.priKey
}

func (a *Account) SetPriKey(priKey *ecdsa.PrivateKey) {
	a.priKey = priKey
}

func (a *Account) Signature(message string) ([]byte, error) {
	return ethutils.Sign([]byte(message), a.priKey)
}

func (a *Account) SignatureFlashBot(message []byte) ([]byte, error) {
	hashedBody := crypto.Keccak256Hash(message).Hex()
	return crypto.Sign(accounts.TextHash([]byte(hashedBody)), a.priKey)
}

func (a *Account) check() bool {
	return len(a.params) == 0 || a.accountTy != KeyStoreTy &&
		a.accountTy != EvMnemonicTy && a.accountTy != EncryptedMnemonicTy &&
		a.accountTy != PlainMnemonicTy && a.accountTy != PlainPrivateKeyTy
}

func NewAccount(_accountType string, _params map[string]interface{}) IAccount {
	return &Account{accountTy: _AccountType(_accountType), params: _params}
}

func getPassPhrase(text string, confirmation bool) string {
	if text != "" {
		fmt.Println(text)
	}
	password, err := prompt.Stdin.PromptPassword("Password: ")
	if err != nil {
		fmt.Printf("Failed to read password: %v", err)
		return ""
	}
	if confirmation {
		confirm, err := prompt.Stdin.PromptPassword("Repeat password: ")
		if err != nil {
			fmt.Printf("Failed to read password confirmation: %v", err)
			return ""
		}
		if password != confirm {
			fmt.Printf("Passwords do not match")
			return ""
		}
	}
	return password
}

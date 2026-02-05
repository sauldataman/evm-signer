package account

import (
	"crypto/ecdsa"
	"evm-signer/pkg/ethutils"
	"evm-signer/pkg/hsm"
	"evm-signer/pkg/hsm/iface"
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
	HSMTy               _AccountType = "HSM"
	MultiHSMTy          _AccountType = "MultiHSM"
)

var (
	errKeyField   = errors.New("account config error, must contains key field")
	errIndexField = errors.New("account config error, must contains index field")

	errKeyArgument              = errors.New("account config error, key must be string")
	errHsmProviderArgument      = errors.New("account config error, provider must be string")
	errHsmPinArgument           = errors.New("account config error, pin code must be string")
	errHsmPublicKeyIdArgument   = errors.New("account config error, public_key_id must be string")
	errHsmPrivateKeyIdArgument  = errors.New("account config error, private_key_id must be string")
	errHsmPrivateKeyIdsArgument = errors.New("account config error, private_key_ids must be string")

	errPassArgument  = errors.New("account config error, pass must be string")
	errUseLastPass   = errors.New("account config error, use_last_pass must be bool")
	errIndexArgument = errors.New("account config error, index must be string, eg: 0-9,256")

	errKeyNull           = errors.New("key field is null")
	errProviderNull      = errors.New("hsm provider field is null")
	errProviderType      = errors.New("at MultiHSM type only supports yubihsm-connector")
	errPrivateKeyIdNull  = errors.New("hsm private_key_id field is null")
	errPrivateKeyIdsNull = errors.New("hsm private_key_ids field is null")
	errIndexNull         = errors.New("index field is null")
)

var logger *logging.SugaredLogger

func SetLogger(_logger *logging.SugaredLogger) {
	logger = _logger
}

type Account struct {
	accountTy  _AccountType
	priKey     *ecdsa.PrivateKey
	client     iface.IHsm
	hsmPbKeyId int64 // HSM 的 public key id
	hsmPvKeyId int64 // HSM 的 private key id
	params     map[string]interface{}
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

	if account.PriKey == nil && account.HsmClient == nil {
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

// key, authKey, pass
func checkHsmParams(params map[string]interface{}) error {
	if _, ok := params["connector_url"]; !ok {
		params["connector_url"] = ""
	}

	// connector_url 为空时，使用默认的 url 进行连接
	if params["connector_url"] == nil {
		params["connector_url"] = ""
	}

	if _, ok := params["provider"]; !ok {
		return errProviderNull
	}

	// provider 不应该为空
	if params["provider"] == nil {
		return errProviderNull
	}

	if reflect.TypeOf(params["provider"]).Kind() != reflect.String {
		return errHsmProviderArgument
	}

	if _, ok := params["pin"]; !ok {
		params["pin"] = ""
	}

	// pin 为空时，需要动态输入 pin code
	if params["pin"] == nil {
		params["pin"] = ""
	}

	if reflect.TypeOf(params["pin"]).Kind() != reflect.String {
		return errHsmPinArgument
	}

	// public_key_id 为选填字段，设置默认值为 -1
	if _, ok := params["public_key_id"]; !ok {
		params["public_key_id"] = -1
	}

	// pin 不应该为空
	if params["public_key_id"] == nil {
		params["public_key_id"] = -1
	}

	if reflect.TypeOf(params["public_key_id"]).Kind() != reflect.Int {
		return errHsmPublicKeyIdArgument
	}

	if _, ok := params["private_key_id"]; !ok {
		return errPrivateKeyIdNull
	}

	// private_key_id 不应该为空
	if params["private_key_id"] == nil {
		return errPrivateKeyIdNull
	}

	if reflect.TypeOf(params["private_key_id"]).Kind() != reflect.Int {
		return errHsmPrivateKeyIdArgument
	}
	return nil
}

func checkMultiHsmParams(params map[string]interface{}) error {
	//1type: MultiHSM
	//provider: yubihsm-connector
	//connector_url: localhost:12346
	//private_key_ids: 1001-1003
	//由于 aws hsm 的 public_key_id 和 private_key_id 要求一一对应，为了简化，此处不考虑 aws hsm。

	if _, ok := params["connector_url"]; !ok {
		params["connector_url"] = ""
	}

	// connector_url 为空时，使用默认的 url 进行连接
	if params["connector_url"] == nil {
		params["connector_url"] = ""
	}

	if _, ok := params["provider"]; !ok {
		return errProviderNull
	}

	// provider 不应该为空
	if params["provider"] == nil {
		return errProviderNull
	}

	if reflect.TypeOf(params["provider"]).Kind() != reflect.String {
		return errHsmProviderArgument
	}

	// provider 不为 yubikey 时返回错误
	if strings.ToLower(params["provider"].(string)) != string(hsm.ProviderYuBiHsm) {
		return errProviderType
	}

	if _, ok := params["pin"]; !ok {
		params["pin"] = ""
	}

	// pin 为空时，需要动态输入 pin code
	if params["pin"] == nil {
		params["pin"] = ""
	}

	if reflect.TypeOf(params["pin"]).Kind() != reflect.String {
		return errHsmPinArgument
	}

	// 解析 private_key_ids
	if _, ok := params["private_key_ids"]; !ok {
		return errPrivateKeyIdsNull
	}

	// private_key_id 不应该为空
	if params["private_key_ids"] == nil {
		return errPrivateKeyIdsNull
	}

	// 字符串，需要拆分，拆分方式和助记词类似。
	if reflect.TypeOf(params["private_key_ids"]).Kind() != reflect.String {
		return errHsmPrivateKeyIdsArgument
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
	case HSMTy:
		params := c.params.(map[string]interface{})
		err := checkHsmParams(params)
		if err != nil {
			return nil, err
		}
		_hsmParams := unWrapHsmData(params)
		_hsmParams.pin = passPhrase(HSMTy, _hsmParams.pin)

		logger.Debugf("_hsmParams.url: %s \n", _hsmParams.url)
		hsmParam := &hsm.ParamsHsm{
			Url:          _hsmParams.url,
			Provider:     hsm.ProviderHsm(_hsmParams.provider),
			LibPath:      "",
			Pin:          _hsmParams.pin,
			PublicKeyId:  _hsmParams.publicKeyId,
			PrivateKeyId: _hsmParams.privateKeyId,
		}
		return NewHsm(hsmParam)
	case MultiHSMTy:
		params := c.params.(map[string]interface{})
		err := checkMultiHsmParams(params)
		if err != nil {
			return nil, err
		}
		_hsmParams := unWrapMultiHsmData(params)
		_hsmParams.pin = passPhrase(MultiHSMTy, _hsmParams.pin)
		logger.Debugf("_hsmParams.url: %s \n", _hsmParams.url)

		hsmParam := &hsm.ParamsHsm{
			Url:          _hsmParams.url,
			Provider:     hsm.ProviderHsm(_hsmParams.provider),
			LibPath:      "",
			Pin:          _hsmParams.pin,
			PublicKeyId:  DefaultKeyId,
			PrivateKeyId: DefaultKeyId,
		}
		return NewMultiYubiHsm(hsmParam, _hsmParams.privateKeyIds)
	default:
		return nil, fmt.Errorf("unSupported account type, only support Keystore, EvMnemonic, " +
			"EncryptedMnemonic, PlainMnemonic, PlainPrivateKey, HSM, MultiHSM at the moment")
	}
}

func passPhrase(accountTy _AccountType, pass string) string {
	if pass == "" {
		password := getPassPhrase(fmt.Sprintf(
			"please enter pin code for [%s] account",
			accountTy,
		), false)
		if password == "" {
			logger.Fatalf("pin code does not empty")
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

func (a *Account) SetHsmClient(client iface.IHsm) {
	a.client = client
}

func (a *Account) SetPublicKeyIdForHsm(keyId int64) {
	a.hsmPbKeyId = keyId
}

func (a *Account) SetPrivateKeyForHsm(keyId int64) {
	a.hsmPvKeyId = keyId
}

func (a *Account) Signature(message string) ([]byte, error) {
	// normal signature
	if a.client == nil {
		return ethutils.Sign([]byte(message), a.priKey)
	}
	// HSM Signature
	a.client.SetPriKeyId(uint(a.hsmPvKeyId))
	return a.client.SignMessage(message)
}

func (a *Account) SignatureFlashBot(message []byte) ([]byte, error) {
	hashedBody := crypto.Keccak256Hash(message).Hex()
	return crypto.Sign(accounts.TextHash([]byte(hashedBody)), a.priKey)
}

func (a *Account) check() bool {
	return len(a.params) == 0 || a.accountTy != KeyStoreTy &&
		a.accountTy != EvMnemonicTy && a.accountTy != EncryptedMnemonicTy &&
		a.accountTy != PlainMnemonicTy && a.accountTy != PlainPrivateKeyTy && a.accountTy != HSMTy
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

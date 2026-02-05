package account

import (
	"cs-evm-signer/types"
	"fmt"
	ethutils "github.com/CoinSummer/go-ethutils"
	dString "github.com/CoinSummer/go-utils/go/string"
	"strconv"
)

type plainMnemonic struct {
	mnemonic string
	indexMap map[int64]struct{}
}

func NewPlainMnemonic(key, indexRange string) (*plainMnemonic, error) {
	if key == "" || indexRange == "" {
		return nil, fmt.Errorf("mnemonic config error")
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
	return &plainMnemonic{mnemonic: key, indexMap: indexMap}, nil
}

func (p *plainMnemonic) Decrypt() []*types.Account {
	var accounts []*types.Account
	for k := range p.indexMap {
		account := &types.Account{
			Index: k,
		}

		key := ethutils.GetAccountFromMnemonic(p.mnemonic, int(k))
		account.Address = key.Address
		logger.Debugf("index: [%d], address: [%s] \n", k, account.Address.String())
		account.PriKey = key.PrivateKey

		accounts = append(accounts, account)
	}
	return accounts
}

func (p *plainMnemonic) Crypto() error {
	return fmt.Errorf("unSupport crypto")
}

type plainPrivateKey struct {
	key string
}

func (p *plainPrivateKey) Decrypt() []*types.Account {
	var accounts []*types.Account

	account := ethutils.GetAccountFromPStr(p.key)
	if account == nil {
		logger.Errorf("plain privateKey is invalidate")
		return nil
	}

	logger.Infof("plain privateKey address: [%s]", account.Address)
	_account := &types.Account{
		Address: account.Address,
		PriKey:  account.PrivateKey,
	}
	accounts = append(accounts, _account)
	return accounts
}

func (p *plainPrivateKey) Crypto() error {
	return fmt.Errorf("unSupport crypto")
}

func NewPlainPrivateKey(key string) (*plainPrivateKey, error) {
	if key == "" {
		return nil, fmt.Errorf("plainPrivateKey config error, private key should't null")
	}
	return &plainPrivateKey{key: key}, nil
}

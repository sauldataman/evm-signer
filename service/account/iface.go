package account

import (
	"crypto/ecdsa"
	"cs-evm-signer/pkg/hsm/iface"
	"cs-evm-signer/types"
)

type (
	// IAccount abstract account
	IAccount interface {
		Account() IAccountOpt
		GetPriKey() *ecdsa.PrivateKey
		SetPriKey(*ecdsa.PrivateKey)
		SetHsmClient(iface.IHsm)
		SetPublicKeyIdForHsm(int64)
		SetPrivateKeyForHsm(int64)
		Signature(message string) ([]byte, error)
		SignatureFlashBot(message []byte) ([]byte, error)
	}

	IAccountOpt interface {
		GetAccount() IGetAccount
		Crypto() (ICrypto, error)
	}

	IGetAccount interface {
		Index(accountMap map[int64]*types.Account, index int64) (*types.Account, bool)
		Address(accountMap map[string]*types.Account, address string) (*types.Account, bool)
	}

	ICrypto interface {
		Decrypt() []*types.Account
		Crypto() error
	}
)

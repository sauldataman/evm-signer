package iface

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type IHsm interface {
	GetUrl() string
	SetUrl(url string)
	GetPubKeyId() uint
	SetPubKeyId(keyId uint)
	GetPriKeyId() uint
	SetPriKeyId(keyId uint)
	GetPublicKey() (publicKey *ecdsa.PublicKey, address common.Address, err error)
	SignTx(chainId *big.Int, tx *types.Transaction) ([]byte, *types.Transaction, error)
	SignEip712(hash []byte) ([]byte, error)
	SignMessage(msg string) ([]byte, error)
}

package _interface

import ethTypes "github.com/ethereum/go-ethereum/core/types"

type IChain interface {
	NewTx(txMsg string) (data *ethTypes.Transaction, err error)
	Sign(string2 string) (string, error)
	Sign712(hash []byte) (string, error)
}

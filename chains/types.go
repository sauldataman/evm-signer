package chains

import "fmt"

type ChainTy string

const (
	EthereumTy ChainTy = "ethereum"
)

var (
	ErrUnSupportedChain = fmt.Errorf("unsupported chain")
)

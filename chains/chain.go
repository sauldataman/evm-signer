package chains

import (
	"cs-evm-signer/chains/ethereum"
	_interface "cs-evm-signer/chains/interface"
)

func GetChain(chainId uint64, chainTy, priKey string) (_interface.IChain, error) {
	switch ChainTy(chainTy) {
	case EthereumTy:
		return ethereum.NewEthChain(chainId, priKey), nil
	default:
		return nil, ErrUnSupportedChain
	}
}

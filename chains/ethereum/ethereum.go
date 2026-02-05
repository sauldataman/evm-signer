package ethereum

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

type EthChain struct {
	chainId uint64
	priKey  string
}

func NewEthChain(chainId uint64, pri string) *EthChain {
	return &EthChain{
		chainId: chainId,
		priKey:  pri,
	}
}

func (ec *EthChain) NewTx(txMsg string) (*ethTypes.Transaction, error) {
	tx := &ethTypes.Transaction{}
	err := json.Unmarshal([]byte(txMsg), tx)
	if err != nil {
		return nil, fmt.Errorf("unmarshal tx error: %s", err)
	}
	return tx, nil
}

func (ec *EthChain) Sign(txMsg string) (string, error) {
	tx, err := ec.NewTx(txMsg)
	if err != nil {
		return "", fmt.Errorf("unmarshal tx error: %s", err)
	}

	signer := ethTypes.LatestSignerForChainID(big.NewInt(int64(ec.chainId)))
	pri, err := crypto.HexToECDSA(ec.priKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key error: %s", err)
	}
	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), pri)
	if err != nil {
		return "", fmt.Errorf("signature error: %s", err)
	}
	return hexutil.Encode(signature), nil
}

func (ec *EthChain) Sign712(hash []byte) (string, error) {
	pri, err := crypto.HexToECDSA(ec.priKey)
	if err != nil {
		return "", err
	}
	signature, err := crypto.Sign(hash, pri)
	if err == nil {
		signature[64] += 27
	}
	return hexutil.Encode(signature), err
}

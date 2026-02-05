package types

import (
	"crypto/ecdsa"
	"evm-signer/pkg/hsm/iface"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Config struct {
	Addr           string `mapstructure:"addr"`
	Port           int    `mapstructure:"port"`
	SSLEnable      bool   `mapstructure:"ssl_enable"`
	SSLCertPath    string `mapstructure:"ssl_cert_path"`
	SSLCertKeyPath string `mapstructure:"ssl_cert_key_path"`
}

type Account struct {
	HsmClient iface.IHsm
	HsmObjId  int64
	Index     int64 // 虚拟助记词的 map id
	Address   common.Address
	PriKey    *ecdsa.PrivateKey
}

type Data struct {
	Data string `json:"data"`
}

type Sign struct {
	Signature string `json:"signature"`
	TxData    string `json:"tx"`
	TxHex     string `json:"tx_hex"`
}

type SignRequest struct {
	Data string `form:"data" binding:"required"`
}

type SignatureMsgInfo struct {
	ChainId int64  `json:"chain_id"`
	Account string `json:"account"`
	Message string `json:"message"`
}

type MevSignatureInfo struct {
	ChainId int64  `json:"chain_id"`
	Index   int64  `json:"index"`
	Data    string `json:"data"`
}

type FlashBotData struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type AddressMsgInfo struct {
	ChainId int64 `json:"chain_id"`
	Index   int64 `json:"index"`
}

type Sign712MsgInfo struct {
	ChainId int64  `json:"chain_id"`
	Account string `json:"account"` // account address
	Data    string `json:"Data"`
}

type MsgInfo struct {
	ChainId     int64  `json:"chain_id"`
	Account     string `json:"account"` // also is address
	Transaction string `json:"transaction"`
}

type Transaction struct {
	TxType               uint8            `json:"-"`
	ChainId              string           `json:"chainId"`
	Type                 string           `json:"type"` // type of transaction, e.g. "legacy, accessList, dynamicFee"
	Hash                 string           `json:"hash,omitempty"`
	Nonce                string           `json:"nonce"`
	From                 string           `json:"from"`
	To                   string           `json:"to"`
	Value                string           `json:"value"`
	Gas                  string           `json:"gas"` // alias gasLimit
	GasPrice             string           `json:"gasPrice"`
	MaxPriorityFeePerGas string           `json:"maxPriorityFeePerGas"` // a.k.a. maxPriorityFeePerGas
	MaxFeePerGas         string           `json:"maxFeePerGas"`         // a.k.a. maxFeePerGas
	Input                string           `json:"input"`
	AccessList           types.AccessList `json:"access_list"` // accessList tx or dynamicFee tx
	V                    string           `json:"v"`
	R                    string           `json:"r"`
	S                    string           `json:"s"`
}

type FmtTransaction struct {
	ChainID    *big.Int
	Nonce      uint64
	GasPrice   *big.Int
	GasTipCap  *big.Int
	GasFeeCap  *big.Int
	Gas        uint64
	To         *common.Address
	Value      *big.Int
	Data       []byte
	AccessList types.AccessList
}

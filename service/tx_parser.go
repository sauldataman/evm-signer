package service

import (
	sTypes "cs-evm-signer/types"
	"encoding/json"
	"math/big"
	"strings"
)

func remove0xPrefix(str string) string {
	return strings.Replace(str, "0x", "", 1)
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func bigIntFromStr(val string) *big.Int {
	value := new(big.Int)
	if has0xPrefix(val) {
		value.SetString(remove0xPrefix(val), 16)
	} else {
		value.SetString(val, 10)
	}
	return value
}

func convertToHex(chainId string, tx *sTypes.Transaction) {
	tx.Type = "0x" + bigIntFromStr(tx.Type).Text(16)
	tx.ChainId = "0x" + bigIntFromStr(chainId).Text(16)
	tx.Nonce = "0x" + bigIntFromStr(tx.Nonce).Text(16)
	tx.Gas = "0x" + bigIntFromStr(tx.Gas).Text(16)
	tx.GasPrice = "0x" + bigIntFromStr(tx.GasPrice).Text(16)
	tx.MaxPriorityFeePerGas = "0x" + bigIntFromStr(tx.MaxPriorityFeePerGas).Text(16)
	tx.MaxFeePerGas = "0x" + bigIntFromStr(tx.MaxFeePerGas).Text(16)
	tx.Value = "0x" + bigIntFromStr(tx.Value).Text(16)
}

func txParse(chainId, msgTx string) (string, error) {
	tx := &sTypes.Transaction{}
	err := json.Unmarshal([]byte(msgTx), tx)
	if err != nil {
		return "", err
	}

	convertToHex(chainId, tx)
	marshal, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

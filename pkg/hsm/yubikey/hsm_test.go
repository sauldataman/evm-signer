package yubikey

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"testing"
)

var (
	url      = DefaultURL
	keyId    = uint16(1000)
	authKey  = uint16(2)
	password = "password1"
	to       = common.HexToAddress("0xA51Fc19f0430614F22B9Caf10491298E5D571313")
)

func init() {
}

func TestHsmClient_Signature(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "test signature",
			message: "Sign me!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewHsmClient(url, keyId, authKey, password)
			if err != nil {
				log.Fatalf("new hsm client error: %s", err)
			}
			signature, err := client.SignMessage(tt.message)
			if err != nil {
				log.Fatalf("call signature error: %s", err)
			}
			t.Logf("res: %s", hexutil.Encode(signature))
		})
	}
}

func TestHsmClient(t *testing.T) {
	tests := []struct {
		name string
		rpc  string
		tx   *types.LegacyTx
	}{
		{
			name: "test tx",
			rpc:  "https://goerli.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161",
			tx: &types.LegacyTx{
				Nonce:    0,
				GasPrice: nil,
				Gas:      21000,
				To:       &to,
				Value:    big.NewInt(2000000000),
				Data:     nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewHsmClient(url, keyId, authKey, password)
			if err != nil {
				log.Fatalf("new hsm client error: %s", err)
			}

			_, address, err := client.GetPublicKey()
			if err != nil {
				log.Fatalf("get publick key via hsm error: %s", err)
			}

			ec, err := ethclient.Dial("https://goerli.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161")
			if err != nil {
				log.Fatal(err)
			}
			nonce, err := ec.PendingNonceAt(context.Background(), address)
			if err != nil {
				log.Fatal(err)
			}
			tt.tx.Nonce = nonce

			log.Println("nonce: ", nonce)
			chainID, err := ec.NetworkID(context.Background())
			if err != nil {
				log.Fatal(err)
			}
			log.Println("chain id:", chainID)
			gasPrice, err := ec.SuggestGasPrice(context.Background())
			if err != nil {
				log.Fatalf("get suggest gas price error: %s", err)
			}
			gasPrice = gasPrice.Mul(gasPrice, big.NewInt(3))
			tt.tx.GasPrice = gasPrice

			signedBytes, signedTx, err := client.SignTx(chainID, types.NewTx(tt.tx))
			if err != nil {
				log.Fatalf("sign tx error: %s", err)
			}
			//logrus.Info(signedTx.RawSignatureValues())
			err = ec.SendTransaction(context.Background(), signedTx)
			if err != nil {
				t.Fatalf("send tx error: %s", err)
			}

			t.Logf("signedBytes: %s", hexutil.Encode(signedBytes))

			t.Logf("tx hash: %s", signedTx.Hash())
		})
	}
}

func TestFmtUrl(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			url: "localhost:12345",
		},
		{
			url: "http://localhost:12345",
		},
		{
			url: "https://localhost:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(fmtUrl(url))
		})
	}
}

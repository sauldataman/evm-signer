package main

import (
	"crypto/rand"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

func init() {
	keyCmd.AddCommand(generateCmd)
}

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "generate a public and private key pair.",
	Long: `
Generate a public and private key. 
They are used for encryption and decryption of transaction bodies. 
The Sodium will use the public key to encrypt the transaction, 
and the Signer will use the private key to decrypt the transaction. 
If you lose the private key, please generate a new one and go to sodium to update the public key`,
}

var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "generate public and private key",
	Example: "./signer key generate",
	Run: func(cmd *cobra.Command, args []string) {
		// Use cryptographically secure random bytes
		entropy := make([]byte, 32)
		if _, err := rand.Read(entropy); err != nil {
			logger.Errorf("failed to generate random bytes: %v", err)
			return
		}

		prvKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), entropy)

		fmt.Println("pri key", hexutil.Encode(crypto.FromECDSA(prvKey.ToECDSA()))[2:])
		fmt.Println("pub key", hexutil.Encode(crypto.FromECDSAPub(pubKey.ToECDSA()))[2:])
	},
}

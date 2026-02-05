package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
	"math"
	"time"
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
		randomBytes := make([]byte, 0)
		cpuPercent, err := cpu.Percent(time.Second, false)
		if err != nil || len(cpuPercent) == 0 {
			logger.Errorf("calculates the percentage of cpu used either per CPU or combined. error: %v", err)
			return
		}

		memory, err := mem.VirtualMemory()
		if err != nil {
			logger.Errorf("get VirtualmemoryStat for memory error: %v", err)
			return
		}

		diskStatus, err := disk.Usage("/")
		if err != nil {
			logger.Errorf("get usage for disk error: %v", err)
			return
		}

		float64ToByte := func(float float64) []byte {
			bits := math.Float64bits(float)
			bytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(bytes, bits)
			return bytes
		}

		randomBytes = append(randomBytes, float64ToByte(cpuPercent[0])...)
		randomBytes = append(randomBytes, float64ToByte(memory.UsedPercent)...)
		randomBytes = append(randomBytes, float64ToByte(diskStatus.UsedPercent)...)

		entropy := sha256.Sum256(randomBytes)
		prvKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), entropy[:])

		fmt.Println("pri key", hexutil.Encode(crypto.FromECDSA(prvKey.ToECDSA()))[2:])
		fmt.Println("pub key", hexutil.Encode(crypto.FromECDSAPub(pubKey.ToECDSA()))[2:])
	},
}

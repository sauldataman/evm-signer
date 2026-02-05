package aws

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"testing"
)

func TestHsmClient_GetPublicKey(t *testing.T) {
	hsm, err := NewAwsHsm("pincode", "", 262159, 262158)
	if err != nil {
		panic(err)
	}
	_, address, err := hsm.GetPublicKey()
	fmt.Println("addr: ", address.String())

	message, err := hsm.SignMessage("hello signer!")
	if err != nil {
		panic(err)
	}
	fmt.Println("sign message: ", hexutil.Encode(message))
}

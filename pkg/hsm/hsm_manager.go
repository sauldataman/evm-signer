package hsm

import (
	"cs-evm-signer/pkg/hsm/aws"
	"cs-evm-signer/pkg/hsm/iface"
	"cs-evm-signer/pkg/hsm/yubikey"
	"fmt"
	"strconv"
	"strings"
)

type ProviderHsm string

const (
	ProviderAws     ProviderHsm = "aws"
	ProviderYuBiHsm ProviderHsm = "yubihsm-connector"
)

type ParamsHsm struct {
	Url          string
	Provider     ProviderHsm
	LibPath      string
	Pin          string
	PublicKeyId  int
	PrivateKeyId int
}

func GetHsmClient(hsmParam *ParamsHsm) (iface.IHsm, error) {
	switch hsmParam.Provider {
	case ProviderAws:
		return aws.NewAwsHsm(hsmParam.Pin, hsmParam.LibPath, hsmParam.PublicKeyId, hsmParam.PrivateKeyId)
	case ProviderYuBiHsm:
		//split pin by 1:password1
		pinArr := strings.Split(hsmParam.Pin, ":")
		if len(pinArr) < 2 {
			return nil, fmt.Errorf("invalid pin length: %d", len(pinArr))
		}

		authId, err := strconv.ParseUint(pinArr[0], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed authId convert to uint16, plz check pin field in yubikey config. error: %s", err)
		}
		return yubikey.NewHsmClient(hsmParam.Url, uint16(hsmParam.PrivateKeyId), uint16(authId), pinArr[1])
	default:
		return nil, fmt.Errorf("[%s] unSupported hsm provider", hsmParam.Provider)
	}
}

package account

import (
	"evm-signer/pkg/hsm"
	"evm-signer/pkg/hsm/iface"
	"evm-signer/types"
	"fmt"
)

type hsmHandle struct {
	client iface.IHsm
}

type hsmParams struct {
	url          string
	provider     string
	pin          string
	publicKeyId  int
	privateKeyId int
}

func unWrapHsmData(hsmData map[string]interface{}) *hsmParams {
	url := hsmData["connector_url"].(string)
	provider := hsmData["provider"].(string)
	pinCode := hsmData["pin"].(string)
	publicKeyId := hsmData["public_key_id"].(int)
	privateKeyId := hsmData["private_key_id"].(int)
	return &hsmParams{
		url:          url,
		provider:     provider,
		pin:          pinCode,
		publicKeyId:  publicKeyId,
		privateKeyId: privateKeyId,
	}
}

func NewHsm(params *hsm.ParamsHsm) (*hsmHandle, error) {
	client, err := hsm.GetHsmClient(params)
	if err != nil {
		return nil, err
	}
	return &hsmHandle{client: client}, nil
}

func (h *hsmHandle) Decrypt() []*types.Account {
	_, address, err := h.client.GetPublicKey()
	if err != nil {
		logger.Fatalf("get public key for yubi error: %s", err)
	}
	var accounts []*types.Account
	logger.Infof("hsm address: [%s] \n", address)
	_account := &types.Account{
		HsmClient: h.client,
		HsmObjId:  int64(h.client.GetPriKeyId()),
		Index:     0,
		Address:   address,
	}
	accounts = append(accounts, _account)
	return accounts
}

func (h *hsmHandle) Crypto() error {
	return fmt.Errorf("unSupport crypto")
}

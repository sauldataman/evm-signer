package account

import (
	"cs-evm-signer/pkg/hsm"
	"cs-evm-signer/pkg/hsm/iface"
	"cs-evm-signer/types"
	"fmt"
	dString "github.com/CoinSummer/go-utils/go/string"
	"sort"
	"strconv"
)

const DefaultKeyId = 0

type multiHsmHandle struct {
	client   iface.IHsm
	indexMap map[int64]struct{}
}

type multiHsmParams struct {
	url           string
	provider      string
	pin           string
	privateKeyIds string
}

func unWrapMultiHsmData(multiHsmData map[string]interface{}) *multiHsmParams {
	url := multiHsmData["connector_url"].(string)
	provider := multiHsmData["provider"].(string)
	pinCode := multiHsmData["pin"].(string)
	privateKeyIds := multiHsmData["private_key_ids"].(string)

	return &multiHsmParams{
		url:           url,
		provider:      provider,
		pin:           pinCode,
		privateKeyIds: privateKeyIds,
	}
}

func NewMultiYubiHsm(params *hsm.ParamsHsm, privateKeyIds string) (*multiHsmHandle, error) {
	client, err := hsm.GetHsmClient(params)
	if err != nil {
		return nil, err
	}

	// 解析 private_key_ids, - 表示连号，前闭后开区间；, 表示独立的 privateKeyId
	indexMap := make(map[int64]struct{})
	splitMap, err := dString.SplitNum(privateKeyIds)
	if err != nil {
		return nil, fmt.Errorf("index config error")
	}

	for _index := range splitMap {
		__index, _ := strconv.ParseInt(_index, 10, 64)
		indexMap[__index] = struct{}{}
	}
	return &multiHsmHandle{client: client, indexMap: indexMap}, nil
}

func (ms *multiHsmHandle) Decrypt() []*types.Account {
	var accounts []*types.Account

	// index map 顺序解析
	sortedKeys := sortMapByKey(ms.indexMap)
	for _, k := range sortedKeys {
		account := &types.Account{
			HsmClient: ms.client,
			HsmObjId:  k,
			Index:     k,
		}

		ms.client.SetPriKeyId(uint(account.HsmObjId))
		_, address, err := ms.client.GetPublicKey()
		if err != nil {
			logger.Fatalf("get public key for [%d] yubi error: %s", account.HsmObjId, err)
		}
		account.Address = address
		logger.Infof("index: [%d], address: [%s] \n", account.HsmObjId, account.Address.String())
		accounts = append(accounts, account)
	}
	return accounts
}

func sortMapByKey(m map[int64]struct{}) []int64 {
	var keys []int64
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func (ms *multiHsmHandle) Crypto() error {
	return fmt.Errorf("unSupport crypto")
}

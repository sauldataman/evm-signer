package service

import (
	"encoding/json"
	"evm-signer/base"
	"evm-signer/service/account"
	"evm-signer/service/rules"
	"evm-signer/types"
	"strings"
)

// GetAccount
// accountForNameMap  [name:index]*types.Account
// accountForIndexMap [index]*types.Account
func GetAccount(scfg *base.SignerConfig) (map[string]*types.Account, map[int64]*types.Account, account.IAccount) {
	accountForAddrMap := make(map[string]*types.Account)
	accountForIndexMap := make(map[int64]*types.Account)
	var insAccount account.IAccount

	accountInfo := scfg.Config.GetStringMap("account")
	if _, ok := accountInfo["type"]; !ok {
		logger.Fatalf("account config error, account.type field not exists.")
	}

	insAccount = account.NewAccount(accountInfo["type"].(string), accountInfo)
	crypto, err := insAccount.Account().Crypto()
	if err != nil {
		logger.Fatalf("invalid account config: %s", err)
	}

	accounts := crypto.Decrypt()
	for _, _account := range accounts {
		accountForAddrMap[strings.ToLower(_account.Address.Hex())] = _account
		accountForIndexMap[_account.Index] = _account
	}
	return accountForAddrMap, accountForIndexMap, insAccount
}

func GetChain(scfg *base.SignerConfig) (map[uint64]*ChainConfig, error) {
	chainMap := make(map[uint64]*ChainConfig)
	chainsMap := scfg.Config.GetStringMap("chains")
	for chainName := range chainsMap {
		chain := new(ChainConfig)
		if err := scfg.Config.UnmarshalKey("chains."+chainName, chain); err != nil {
			logger.Fatalf("invalid chain config: %s", err)
		}
		if chain.ChainId == 0 {
			logger.Fatalf("invalid chain config, chainId should't be 0")
		}

		if chain.ChainType == "" {
			logger.Fatalf("invalid chain config, chain type can't be null")
		}

		chain.Name = chainName
		chain.ChainType = strings.ToLower(chain.ChainType)
		chainMap[chain.ChainId] = chain
	}

	return chainMap, nil
}

type AuthConfig struct {
	IP string `mapstructure:"ip"`
}

func GetAuthConfig(scfg *base.SignerConfig) *AuthConfig {
	signerCnf := new(AuthConfig)
	if err := scfg.Config.UnmarshalKey("auth", signerCnf); err != nil {
		logger.Fatalf("invalid signer config: %s", err)
		return nil
	}
	return signerCnf
}

func GetHttpConfig(scfg *base.SignerConfig) *types.Config {
	httpConfig := new(types.Config)
	if err := scfg.Config.UnmarshalKey("listen", &httpConfig); err != nil {
		logger.Fatalf("invalid listen config:%s", err)
	}
	return httpConfig
}

func GetIpWhiteList(ipList []string) map[string]struct{} {
	ipWhiteList := make(map[string]struct{})
	for _, ip := range ipList {
		if ip == "" {
			continue
		} else {
			ipWhiteList[ip] = struct{}{}
		}
	}
	return ipWhiteList
}

func GetRuleConfig(scfg *base.SignerConfig) (rules.Rules, error) {
	_rules := new(rules.Rules)
	err := json.Unmarshal(scfg.Rule, _rules)
	if err != nil {
		return nil, err
	}
	return *_rules, nil
}

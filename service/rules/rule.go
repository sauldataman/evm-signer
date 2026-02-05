package rules

import (
	"cs-evm-signer/types"
	"github.com/CoinSummer/go-base/logging"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"strings"
)

type Rules []*Rule

var logger *logging.SugaredLogger

func SetLogger(_logger *logging.SugaredLogger) {
	logger = _logger
}

func (c Rules) Length() int {
	return len(c)
}

func (c Rules) GetMatched(chainId int64, tx *types.Transaction) *Rule {
	for _, rule := range c {
		isMatch := rule.IsMatch(chainId, tx)
		if isMatch {
			return rule
		}
	}
	return nil
}

func (c Rules) GetMatchedEip712(chainId int64, eip712Msg *apitypes.TypedData) *Rule {
	for _, rule := range c {
		isMatch := rule.IsMatch712(chainId, eip712Msg)
		if isMatch {
			return rule
		}
		logger.Infof("[RuleNotMatch] %s", rule.Name)
	}
	return nil
}

// GetMatchedMessage message 是否和 rule 中的关键词匹配
// 区分大小写
func (c Rules) GetMatchedMessage(chainId int64, message string) *Rule {
	for _, rule := range c {
		isMatch := rule.IsMatchMessage(chainId, message)
		if isMatch {
			return rule
		}
	}
	return nil
}

func (c Rules) Init() {
	for _, rule := range c {
		rule.Init()
	}
}

type Rule struct {
	Name       string      `json:"name" mapstructure:"name"`
	ChainId    int64       `json:"chain_id" mapstructure:"chain_id"`
	Conditions *Conditions `json:"conditions" mapstructure:"conditions"`
}

func (r *Rule) IsMatch(chainId int64, tx *types.Transaction) bool {
	if r.ChainId != chainId {
		return false
	}
	tx.From = strings.ToLower(tx.From)
	tx.To = strings.ToLower(tx.To)
	if !r.Conditions.IsMatch(tx) {
		return false
	}
	return true
}

func (r *Rule) IsMatch712(chainId int64, eip712Msg *apitypes.TypedData) bool {
	if r.ChainId != chainId {
		return false
	}

	if !r.Conditions.IsMatch712(eip712Msg) {
		return false
	}
	return true
}

func (r *Rule) IsMatchMessage(chainId int64, message string) bool {
	if r.ChainId != chainId {
		return false
	}

	if !r.Conditions.IsMatchMessage(strings.ToLower(message)) {
		return false
	}
	return true
}

func (r *Rule) Init() {
	r.Conditions.Init()
}

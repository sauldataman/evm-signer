package rules

import (
	"evm-signer/types"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/shopspring/decimal"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type Field string

const (
	FromField                     Field = "from"
	ToField                       Field = "to"
	ValueField                    Field = "value"
	DataSelectorField             Field = "data_selector"
	DataField                     Field = "data"
	DataParamField                Field = "data_param"
	MessageField                  Field = "message"
	Eip712DomainName              Field = "eip712.domain.name"
	Eip712DomainVersion           Field = "eip712.domain.version"
	Eip712DomainChainId           Field = "eip712.domain.chainId"
	Eip712DomainVerifyingContract Field = "eip712.domain.verifyingContract"
	Eip712PrimaryType             Field = "eip712.primaryType"
)

func (f Field) IsValid() bool {
	return string(f) != ""
}

type Symbol string

const (
	EqualSymbol         Symbol = "=="
	GrateAndEqualSymbol Symbol = ">="
	InSymbol            Symbol = "in"
	LessAndEqualSymbol  Symbol = "<="
	ContainsSymbol      Symbol = "contains"
	RegexSymbol         Symbol = "regex"
)

func (f Symbol) IsValid() bool {
	return string(f) != ""
}

type Conditions []*Condition

func (c Conditions) Init() {
	for _, con := range c {
		con.Init()
	}
}

// IsMatch MUST match all
func (c Conditions) IsMatch(tx *types.Transaction) bool {
	for _, condition := range c {
		if !condition.IsMatch(tx) {
			return false
		}
	}
	return true
}

func (c Conditions) IsMatch712(eip712Msg *apitypes.TypedData) bool {
	for _, condition := range c {
		if !condition.IsMatch712(eip712Msg) {
			return false
		}
	}
	return true
}

func (c Conditions) IsMatchMessage(message string) bool {
	for _, condition := range c {
		if !condition.IsMatchMessage(message) {
			return false
		}
	}
	return true
}

type Condition struct {
	Field    Field  `json:"field"`
	Symbol   Symbol `json:"symbol"`
	Value    string `json:"value"`
	Abi      string `json:"abi"`
	Param    string `json:"param"` // 自定义ABI的比较参数名
	inputs   abi.Arguments
	selector string
}

func (c *Condition) Init() {
	// lowerCase
	c.Value = strings.ToLower(c.Value)
	// init abi
	if c.Abi == "" {
		return
	}
	_abi, err := abi.JSON(strings.NewReader("[" + c.Abi + "]"))
	if err != nil {
		logger.Warnf("can not parser abi files %s", err.Error())
		return
	}
	var funcName string
	for funcName, _ = range _abi.Methods {
		break
	}
	if _, ok := _abi.Methods[funcName]; !ok {
		logger.Warnf("abi not contains func: %s", c.Param)
		return
	}
	c.selector = hexutil.Encode(_abi.Methods[funcName].ID)
	c.inputs = _abi.Methods[funcName].Inputs
}

func (c *Condition) IsMatch712(msg712 *apitypes.TypedData) bool {
	isMatch := false
	switch c.Field {
	case Eip712DomainName:
		isMatch = c.IsMatchString(strings.ToLower(msg712.Domain.Name), c.Symbol)
		if !isMatch {
			logger.Warnf("[ConditionMisMatch] eip721.domain.name is [ %s ] != [ %s ]", c.Value, msg712.Domain.Name)
		}
		return isMatch
	case Eip712DomainVersion:
		isMatch = c.IsMatchString(msg712.Domain.Version, c.Symbol)
		if !isMatch {
			logger.Warnf("[ConditionMisMatch] eip721.domain.version is %s != %s", c.Value, msg712.Domain.Version)
		}
		return isMatch
	case Eip712DomainChainId:
		isMatch = c.IsMatchBigInt((*big.Int)(msg712.Domain.ChainId), c.Symbol)
		if !isMatch {
			chainIdTxt, _ := msg712.Domain.ChainId.MarshalText()
			parseInt, _ := strconv.ParseInt(string(chainIdTxt)[2:], 16, 32)
			logger.Warnf("[ConditionMisMatch] eip721.domain.chainId is %s != %d", c.Value, parseInt)
		}
		return isMatch
	case Eip712DomainVerifyingContract:
		isMatch = c.IsMatchString(msg712.Domain.VerifyingContract, c.Symbol)
		if !isMatch {
			logger.Warnf("[ConditionMisMatch] eip721.domain.VerifyingContract is %s != %s", c.Value, msg712.Domain.VerifyingContract)
		}
		return isMatch
	case Eip712PrimaryType:
		isMatch = c.IsMatchString(msg712.PrimaryType, c.Symbol)
		if !isMatch {
			logger.Warnf("[ConditionMisMatch] eip712.primaryType is %s != %s", c.Value, msg712.PrimaryType)
		}
		return isMatch
	default:
		conditionFields := strings.Split(string(c.Field), ".")
		if len(conditionFields) >= 2 {
			//check condition field
			if conditionField := conditionFields[1]; conditionField == "message" {
				column := conditionFields[2]
				primaryTypes := msg712.Types[msg712.PrimaryType]
				for _, t := range primaryTypes {
					if t.Name == column {
						columnType := t.Type
						if t.Type[:4] == "uint" {
							columnType = "uint"
						}
						switch columnType {
						case "address":
							isMatch = c.IsMatchString(msg712.Message[t.Name].(string), c.Symbol)
							if !isMatch {
								logger.Warnf("[ConditionMisMatch] eip721.message.%s is %s != %s", t.Name, c.Value, msg712.Message[t.Name].(string))
							}
							return isMatch
						case "string":
							isMatch = c.IsMatchString(msg712.Message[t.Name].(string), c.Symbol)
							if !isMatch {
								logger.Warnf("[ConditionMisMatch] eip721.message.%s is %s != %s", t.Name, c.Value, msg712.Message[t.Name].(string))
							}
							return isMatch
						case "bytes":
							isMatch = c.IsMatchString(msg712.Message[t.Name].(string), c.Symbol)
							if !isMatch {
								logger.Warnf("[ConditionMisMatch] eip721.message.%s is %s != %s", t.Name, c.Value, msg712.Message[t.Name].(string))
							}
							return isMatch
						case "uint":
							uintNumber, _ := decimal.NewFromString(msg712.Message[t.Name].(string))
							isMatch = c.IsMatchBigInt(uintNumber.BigInt(), c.Symbol)
							if !isMatch {
								chainIdTxt, _ := msg712.Domain.ChainId.MarshalText()
								parseInt, _ := strconv.ParseInt(string(chainIdTxt)[2:], 16, 32)
								logger.Warnf("[ConditionMisMatch] eip721.message.%s is %s != %d", t.Name, c.Value, parseInt)
							}
						case "bool":
							isMatch = c.IsMatchBool(msg712.Message[t.Name].(bool), c.Symbol)
							if !isMatch {
								logger.Warnf("[ConditionMisMatch] eip721.message.%s is %s != %s", t.Name, c.Value, msg712.Message[t.Name].(string))
							}
							return isMatch
						}
					}
				}
			}
		}
		return false
	}
}

func (c *Condition) IsMatchMessage(message string) bool {
	switch c.Field {
	case MessageField:
		return c.IsMatchString(message, c.Symbol)
	}
	return false
}

func (c *Condition) IsMatch(tx *types.Transaction) bool {
	switch c.Field {
	case FromField:
		return c.IsMatchString(tx.From, c.Symbol)
	case ToField:
		return c.IsMatchString(tx.To, c.Symbol)
	case ValueField:
		var value *big.Int
		var ok bool
		if strings.HasPrefix(tx.Value, "0x") || strings.HasPrefix(tx.Value, "0X") {
			value, ok = new(big.Int).SetString(tx.Value[2:], 16)
		} else {
			value, ok = new(big.Int).SetString(tx.Value, 10)
		}
		if !ok {
			logger.Warnf("[ValueField] tx.Value can not convert to big.int: %s", tx.Value)
			return false
		}
		return c.IsMatchBigInt(value, c.Symbol)
	case DataSelectorField:
		if len(tx.Input) < 10 {
			return false
		}
		return c.IsMatchString(tx.Input[0:10], c.Symbol)
	case DataField:
		c.Value = strings.ToLower(c.Value)
		return c.IsMatchString(strings.ToLower(tx.Input), c.Symbol)
	case DataParamField:
		return c.isMatchDataParam(tx.Input)
	default:
		return false
	}
}

func (c *Condition) isMatchDataParam(input string) bool {
	if c.inputs == nil || len(c.inputs) == 0 {
		logger.Warnf("[DataParamField] abi not initialized, check abi field in condition")
		return false
	}
	if c.Param == "" {
		logger.Warnf("[DataParamField] param field is empty")
		return false
	}
	if len(input) < 10 {
		logger.Warnf("[DataParamField] input too short: %s", input)
		return false
	}
	selector := strings.ToLower(input[0:10])
	if selector != c.selector {
		logger.Warnf("[DataParamField] selector mismatch: expected %s, got %s", c.selector, selector)
		return false
	}
	data, err := hexutil.Decode(input)
	if err != nil {
		logger.Warnf("[DataParamField] failed to decode input: %s", err.Error())
		return false
	}
	params, err := c.inputs.Unpack(data[4:])
	if err != nil {
		logger.Warnf("[DataParamField] failed to unpack params: %s", err.Error())
		return false
	}
	for i, arg := range c.inputs {
		if arg.Name == c.Param {
			switch v := params[i].(type) {
			case *big.Int:
				return c.IsMatchBigInt(v, c.Symbol)
			case string:
				return c.IsMatchString(v, c.Symbol)
			default:
				logger.Warnf("[DataParamField] unsupported param type: %T", v)
				return false
			}
		}
	}
	logger.Warnf("[DataParamField] param %s not found in abi", c.Param)
	return false
}

func (c *Condition) IsMatchString(value string, symbol Symbol) bool {
	switch symbol {
	case EqualSymbol:
		return strings.EqualFold(value, c.Value)
	case InSymbol:
		_values := strings.Split(c.Value, ",")
		return IsContains(_values, value)
	case ContainsSymbol:
		return strings.Contains(value, c.Value)
	case RegexSymbol:
		matched, err := regexp.MatchString(c.Value, value)
		if err != nil {
			return false
		}
		return matched
	default:
		logger.Warnf("unpported symbol %s", symbol)
		return false
	}
}

func (c *Condition) IsMatchBigInt(value *big.Int, symbol Symbol) bool {
	_value, match := new(big.Int).SetString(c.Value, 10)
	if !match {
		logger.Warnf("[%s] condition value can not convent from string to big.int", c.Value)
		return false
	}
	switch symbol {
	case EqualSymbol:
		return value.Cmp(_value) == 0
	case GrateAndEqualSymbol:
		return value.Cmp(_value) >= 0
	case LessAndEqualSymbol:
		return value.Cmp(_value) <= 0
	default:
		return false
	}
}

func (c *Condition) IsMatchBool(value bool, symbol Symbol) bool {
	switch symbol {
	case EqualSymbol:
		return fmt.Sprintf("%t", value) == c.Value
	default:
		return false
	}
}

func IsContains(arr []string, str string) bool {
	for _, s := range arr {
		if strings.EqualFold(s, str) {
			return true
		}
	}
	return false
}

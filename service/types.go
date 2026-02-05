package service

import (
	"fmt"
)

var (
	ErrUnSupport = fmt.Errorf("unSupport chain")
	ErrIllegalIP = fmt.Errorf("illegal ip request")
)

type ResponseMsg struct {
	Code ErrCode     `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type ErrCode int

const (
	InvalidFormData ErrCode = 4000 + iota
	ChainError
	SignError
	AuthError
	InternalError
	HeaderError
	ExpiredRequest
	IllegalAccess
	IllegalTransaction
	ParseError
	ParamError
	ForbiddenError
)

var ErrorMsgMap = map[ErrCode]string{
	InvalidFormData:    "invalid form data",
	ChainError:         "unsupported chain",
	SignError:          "sign error",
	AuthError:          "auth error",
	InternalError:      "internal error",
	HeaderError:        "invalid header",
	ForbiddenError:     "rule check forbidden",
	ExpiredRequest:     "expired request",
	IllegalAccess:      "illegal access",
	IllegalTransaction: "illegal transaction",
	ParseError:        "parse error",
	ParamError:         "param error",
}

type MyError struct {
	Code ErrCode
}

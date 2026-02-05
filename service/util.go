package service

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

func (e *MyError) Error() string {
	if ErrorMsgMap[e.Code] == "" {
		return "unknown error"
	}
	return ErrorMsgMap[e.Code]
}

func priKey2Str(priKey *ecdsa.PrivateKey) string {
	privateKeyBytes := crypto.FromECDSA(priKey)
	return hexutil.Encode(privateKeyBytes)[2:]
}

func ReturnError(c *gin.Context, code ErrCode, msg string) {
	c.AbortWithStatusJSON(400, ResponseMsg{
		Code: code,
		Msg:  msg,
	})
}

func ReturnSuccess(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(200, ResponseMsg{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

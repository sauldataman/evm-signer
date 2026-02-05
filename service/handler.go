package service

import (
	"evm-signer/chains"
	sTypes "evm-signer/types"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gin-gonic/gin"
	"math/big"
	"strings"
)

// GetSignMessage match rule, must check to
func (s *Service) GetSignMessage(ctx *gin.Context) {
	if err := s.CheckIp(ctx); err != nil {
		_msg := fmt.Sprintf("ip: [ %s ] illegal", ctx.ClientIP())
		ReturnError(ctx, IllegalAccess, _msg)
		return
	}

	msgData, code, err := s.getMsgData(ctx)
	if err != nil {
		_msg := fmt.Sprintf("decode msgData for getSignature error: [ %s ]", err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, code, _msg)
		return
	}

	msgInfo := sTypes.SignatureMsgInfo{}
	// TODO: 特殊字符处理
	fmtData := strings.Replace(string(msgData), "\n", "\\n", -1)

	err = json.Unmarshal([]byte(fmtData), &msgInfo)
	if err != nil {
		_msg := fmt.Sprintf("unmarshal [ %s ] msgData error: [ %s ]", string(msgData), err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, ParamError, _msg)
		return
	}

	if 0 >= msgInfo.ChainId {
		_msg := fmt.Sprintf("chainId: [ %d ] <= 0, chainId should be > 0", msgInfo.ChainId)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	if "" == msgInfo.Message {
		_msg := "message for getSignature is null"
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	if msgInfo.Account == "" {
		_msg := fmt.Sprintf("account is null, plz check your account")
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	ai, ok := s.GetAccount(msgInfo.Account)
	if !ok {
		_msg := fmt.Sprintf("can't matched an account via [ %s ] account for [ %s ] messgae on [ %d ] chain_id",
			msgInfo.Account, msgInfo.Message, msgInfo.ChainId)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	// match rules
	matchRule := s.rules.GetMatchedMessage(msgInfo.ChainId, msgInfo.Message)
	if matchRule == nil {
		_msg := fmt.Sprintf("match rule via sign message was mismatched via [ %d ] chainId, [ %s ] message",
			msgInfo.ChainId, msgInfo.Message)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}
	logger.Infof("request mathed rule [ %s ]", matchRule.Name)

	s.iAccount.SetPriKey(ai.PriKey)
	s.iAccount.SetHsmClient(ai.HsmClient)
	s.iAccount.SetPrivateKeyForHsm(ai.HsmObjId)
	//ai.Index
	signature, err := s.iAccount.Signature(msgInfo.Message)
	if err != nil {
		_msg := fmt.Sprintf("get signature for [ %s ] message on [ %d ] chain error: [ %s ]",
			msgInfo.Message, msgInfo.ChainId, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	data := sTypes.Data{
		Data: hexutil.Encode(signature),
	}

	logger.Infof("request ip: [ %s ], account: [ %s ], chain_id: [ %d ], message: [ %s ], signed message: [ %s ]",
		ctx.ClientIP(), msgInfo.Account, msgInfo.ChainId, msgInfo.Message, data.Data)
	ctx.AbortWithStatusJSON(200, data)
}

// GetAddress match rule, must check to
func (s *Service) GetAddress(ctx *gin.Context) {
	if err := s.CheckIp(ctx); err != nil {
		_msg := fmt.Sprintf("ip: [ %s ] illegal", ctx.ClientIP())
		ReturnError(ctx, IllegalAccess, _msg)
		return
	}

	msgData, code, err := s.getMsgData(ctx)
	if err != nil {
		_msg := fmt.Sprintf("parser msgData for getAddress error: [ %s ]", err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, code, _msg)
		return
	}

	msgInfo := sTypes.AddressMsgInfo{}
	err = json.Unmarshal(msgData, &msgInfo)
	if err != nil {
		_msg := fmt.Sprintf("decode msgData for getAddress error: [ %s ]", err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, ParamError, _msg)
		return
	}

	if 0 >= msgInfo.ChainId {
		_msg := fmt.Sprintf("chainId: [ %d ] <= 0, chainId should be > 0", msgInfo.ChainId)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	if msgInfo.Index < 0 {
		_msg := fmt.Sprintf("account_index: [ %d ] < 0, should be >= 0", msgInfo.Index)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	// return address
	ai, ok := s.GetAccountList(msgInfo.Index)
	if !ok {
		_msg := fmt.Sprintf("can't matched an account via [ %d ] account index on [ %d ] chain id",
			msgInfo.Index, msgInfo.ChainId)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	data := sTypes.Data{
		Data: ai.Address.String(),
	}

	logger.Infof("request ip: [ %s ], chain_id: [ %d ], account index: [ %d ], resp account: [ %s ]",
		ctx.ClientIP(), msgInfo.ChainId, msgInfo.Index, data.Data)
	ctx.AbortWithStatusJSON(200, data)
}

// GetSign712 处理 EIP-712 类型化数据签名请求
// 更多信息：https://eips.ethereum.org/EIPS/eip-712
func (s *Service) GetSign712(ctx *gin.Context) {
	if err := s.CheckIp(ctx); err != nil {
		_msg := fmt.Sprintf("ip: [ %s ] illegal", ctx.ClientIP())
		ReturnError(ctx, IllegalAccess, _msg)
		return
	}

	msgData, code, err := s.getMsgData(ctx)
	if err != nil {
		_msg := fmt.Sprintf("parse msg error: [ %s ]", err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, code, _msg)
		return
	}

	msgInfo := sTypes.Sign712MsgInfo{}
	err = json.Unmarshal(msgData, &msgInfo)
	if err != nil {
		_msg := fmt.Sprintf("unmarshal [ %s ] msg error: [ %s ]", string(msgData), err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, ParamError, err.Error())
		return
	}

	eip712Data := apitypes.TypedData{}
	err = json.Unmarshal([]byte(msgInfo.Data), &eip712Data)
	if err != nil {
		_msg := fmt.Sprintf("unmarshal [ %s ] eip712Data error: [ %s ]", msgInfo.Data, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, ParamError, err.Error())
		return
	}

	if eip712Data.Types == nil {
		_msg := fmt.Sprintf("eip712 types data is null")
		logger.Errorf(_msg)
		ReturnError(ctx, ParamError, err.Error())
		return
	}

	//eip712Data.Types["Root"] = []apitypes.Type{
	//	{Name: "root", Type: "bytes32"},
	//}
	//eip712Data.Types["EIP712Domain"] = []apitypes.Type{
	//	{Name: "name", Type: "string"},
	//	{Name: "version", Type: "string"},
	//	{Name: "chainId", Type: "uint256"},
	//	{Name: "verifyingContract", Type: "address"},
	//}

	if eip712Data.PrimaryType == "" {
		_msg := fmt.Sprintf("primary type is null")
		logger.Errorf(_msg)
		ReturnError(ctx, ParamError, _msg)
		return
	}

	//logger.Infof("msgData: %v", msgInfo.Data)

	if msgInfo.ChainId <= 0 {
		_msg := fmt.Sprintf("chainId: [ %d ] <= 0, chainId should be > 0", msgInfo.ChainId)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	if "" == msgInfo.Account {
		_msg := fmt.Sprintf("account is null")
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	chainConfig := s.GetChainConfig(uint64(msgInfo.ChainId))
	if chainConfig == nil {
		_msg := fmt.Sprintf("chainConfig via chain_id is null")
		logger.Errorf(_msg)
		ReturnError(ctx, ChainError, _msg)
		return
	}

	ai, ok := s.GetAccount(msgInfo.Account)
	if !ok {
		_msg := fmt.Sprintf("[ %s ] account not exist", msgInfo.Account)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	// match rule
	matchRule := s.rules.GetMatchedEip712(msgInfo.ChainId, &eip712Data)
	if matchRule == nil {
		_msg := "match rule via transaction was mismatched"
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}
	logger.Infof("request mathed rule [ %s ]", matchRule.Name)

	hashData, _, err := apitypes.TypedDataAndHash(eip712Data)
	if err != nil {
		_msg := fmt.Sprintf("[ %d ] chain convert params to TypedDataAndHash error: [ %s ]",
			chainConfig.ChainId, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, ParamError, _msg)
		return
	}

	if ai.HsmClient != nil {
		logger.Infof("using hsm sign message")
		// 设置 privateKeyId
		logger.Infof("using [ %d ] hsmObjId, account index [ %d ]", ai.HsmObjId, ai.Index)
		ai.HsmClient.SetPriKeyId(uint(ai.HsmObjId))
		signedBytes, err := ai.HsmClient.SignEip712(hashData)
		if err != nil {
			_msg := fmt.Sprintf("sign tx via hsm for [ %s ] eip712 message error: [ %s ]", msgInfo.Data, err)
			logger.Errorf(_msg)
			ReturnError(ctx, InvalidFormData, _msg)
			return
		}
		sign := sTypes.Sign{
			Signature: hexutil.Encode(signedBytes),
		}

		logger.Infof("[EIP712] request ip: [ %s ], chain_id: [ %d ], account: [ %s ], messgae: [ %s ], signed data: [ %s ]",
			ctx.ClientIP(), msgInfo.ChainId, msgInfo.Account, msgInfo.Data, sign.Signature)
		ctx.AbortWithStatusJSON(200, sign)
		return
	}

	key := priKey2Str(ai.PriKey)
	chain, err := chains.GetChain(chainConfig.ChainId, chainConfig.ChainType, key)
	if err != nil {
		_msg := fmt.Sprintf("[ %d ] chain config find error: [ %s ]", chainConfig.ChainId, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, ChainError, _msg)
		return
	}

	signature, err := chain.Sign712(hashData)
	if err != nil {
		_msg := fmt.Sprintf("get chain sign for [ %s ] transaction error: [ %s ]",
			hexutil.Encode(hashData), err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, SignError, _msg)
		return
	}

	sign := sTypes.Sign{
		Signature: signature,
	}

	logger.Infof("[EIP712] request ip: [ %s ], chain_id: [ %d ], account: [ %s ], messgae: [ %s ], signed data: [ %s ]",
		ctx.ClientIP(), msgInfo.ChainId, msgInfo.Account, msgInfo.Data, sign.Signature)
	ctx.AbortWithStatusJSON(200, sign)
}

// GetSign 处理交易签名请求
func (s *Service) GetSign(ctx *gin.Context) {
	if err := s.CheckIp(ctx); err != nil {
		_msg := fmt.Sprintf("ip: [ %s ] illegal", ctx.ClientIP())
		ReturnError(ctx, IllegalAccess, _msg)
		return
	}
	msgData, code, err := s.getMsgData(ctx)
	if err != nil {
		_msg := fmt.Sprintf("parse msg error: [ %s ]", err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, code, _msg)
		return
	}

	msgInfo := sTypes.MsgInfo{}
	err = json.Unmarshal(msgData, &msgInfo)
	if err != nil {
		_msg := fmt.Sprintf("unmarshal [ %s ] msg error: [ %s ]", string(msgData), err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, ParamError, err.Error())
		return
	}

	if msgInfo.ChainId <= 0 {
		_msg := fmt.Sprintf("chainId: [ %d ] <= 0, chainId should be > 0", msgInfo.ChainId)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	if "" == msgInfo.Account {
		_msg := fmt.Sprintf("account is null")
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	if "" == msgInfo.Transaction {
		_msg := fmt.Sprintf("transaction is null")
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	chainConfig := s.GetChainConfig(uint64(msgInfo.ChainId))
	if chainConfig == nil {
		_msg := fmt.Sprintf("chainConfig via chain_id is null")
		logger.Errorf(_msg)
		ReturnError(ctx, ChainError, _msg)
		return
	}

	ai, ok := s.GetAccount(msgInfo.Account)
	if !ok {
		_msg := fmt.Sprintf("[ %s ] account not exist", msgInfo.Account)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	// parse tx
	tx := new(sTypes.Transaction)
	err = json.Unmarshal([]byte(msgInfo.Transaction), tx)
	if err != nil {
		_msg := fmt.Sprintf("[ %s ] transacton format error: [ %s ]", msgInfo.Transaction, err.Error())
		logger.Warnf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}
	tx.From = strings.ToLower(msgInfo.Account)

	// match rule
	matchRule := s.rules.GetMatched(msgInfo.ChainId, tx)
	if matchRule == nil {
		_msg := fmt.Sprintf("match rule via [ %s ] transaction for [ %s ] account on [ %d ] chainId was mismatched",
			msgInfo.Transaction, msgInfo.Account, msgInfo.ChainId)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}
	logger.Infof("request mathed rule [ %s ]", matchRule.Name)

	// convert
	msgInfo.Transaction, err = txParse(fmt.Sprintf("%d", msgInfo.ChainId), msgInfo.Transaction)
	if err != nil {
		_msg := fmt.Sprintf("txParse transaction was invalid, error: [ %s ]", err)
		logger.Errorf(_msg)
		ReturnError(ctx, InvalidFormData, _msg)
		return
	}

	if ai.HsmClient != nil {
		chain, err := chains.GetChain(chainConfig.ChainId, chainConfig.ChainType, "")
		if err != nil {
			_msg := fmt.Sprintf("[ %d ] chain config find error: [ %s ]", chainConfig.ChainId, err)
			logger.Errorf(_msg)
			ReturnError(ctx, ChainError, _msg)
			return
		}

		txData, err := chain.NewTx(msgInfo.Transaction)
		if err != nil {
			_msg := fmt.Sprintf("[ %d ] chain new tx error: [ %s ]", chainConfig.ChainId, err)
			logger.Errorf(_msg)
			ReturnError(ctx, ChainError, _msg)
			return
		}

		// 设置 privateKeyId
		logger.Infof("using [ %d ] hsmObjId, account index [ %d ]", ai.HsmObjId, ai.Index)
		ai.HsmClient.SetPriKeyId(uint(ai.HsmObjId))
		signedBytes, _, err := ai.HsmClient.SignTx(big.NewInt(int64(chainConfig.ChainId)), txData)
		if err != nil {
			_msg := fmt.Sprintf("sign tx via hsm error: [ %s ]", err)
			logger.Errorf(_msg)
			ReturnError(ctx, InvalidFormData, _msg)
			return
		}

		signer := types.LatestSignerForChainID(big.NewInt(msgInfo.ChainId))
		txData, err = txData.WithSignature(signer, signedBytes)
		if err != nil {
			_msg := fmt.Sprintf("[ %d ] chain call WithSignature error: [ %s ]", msgInfo.ChainId, err.Error())
			logger.Errorf(_msg)
			ReturnError(ctx, SignError, _msg)
			return
		}
		marshalJSON, err := txData.MarshalJSON()
		if err != nil {
			_msg := fmt.Sprintf("[ %d ] chain call MarshalJSON error: [ %s ]", msgInfo.ChainId, err.Error())
			logger.Errorf(_msg)
			ReturnError(ctx, SignError, _msg)
			return
		}

		txHex, err := txData.MarshalBinary()
		if err != nil {
			_msg := fmt.Sprintf("[ %d ] chain call MarshalBinary error: [ %s ]", msgInfo.ChainId, err.Error())
			logger.Errorf(_msg)
			ReturnError(ctx, SignError, _msg)
			return
		}

		sign := sTypes.Sign{
			Signature: hexutil.Encode(signedBytes),
			TxData:    string(marshalJSON),
			TxHex:     hexutil.Encode(txHex),
		}

		logger.Infof("[Sign Transaction] request ip: [ %s ], chain_id: [ %d ], account: [ %s ], Transaction: [ %s ], signed data: [ %s ]",
			ctx.ClientIP(), msgInfo.ChainId, msgInfo.Account, msgInfo.Transaction, sign.Signature)
		ctx.AbortWithStatusJSON(200, sign)
		return
	}

	key := priKey2Str(ai.PriKey)
	chain, err := chains.GetChain(chainConfig.ChainId, chainConfig.ChainType, key)
	if err != nil {
		_msg := fmt.Sprintf("[ %d ] chain config find error: [ %s ]", chainConfig.ChainId, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, ChainError, _msg)
		return
	}

	// normal sign or hsm sign
	signature, err := chain.Sign(msgInfo.Transaction)
	if err != nil {
		_msg := fmt.Sprintf("get chain sign for [ %s ] transaction error: [ %s ]", tx.Hash, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, SignError, _msg)
		return
	}

	txData, err := chain.NewTx(msgInfo.Transaction)
	if err != nil {
		_msg := fmt.Sprintf("[ %d ] chain new tx error: [ %s ]", msgInfo.ChainId, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, SignError, _msg)
		return
	}

	signatureBytes, err := hexutil.Decode(signature)
	if err != nil {
		_msg := fmt.Sprintf("[ %s ] signature decode error: [ %s ]", signature, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, SignError, _msg)
		return
	}

	signer := types.LatestSignerForChainID(big.NewInt(msgInfo.ChainId))
	txData, err = txData.WithSignature(signer, signatureBytes)
	if err != nil {
		_msg := fmt.Sprintf("[ %d ] chain call WithSignature error: [ %s ]", msgInfo.ChainId, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, SignError, _msg)
		return
	}
	marshalJSON, err := txData.MarshalJSON()
	if err != nil {
		_msg := fmt.Sprintf("[ %d ] chain call MarshalJSON error: [ %s ]s", msgInfo.ChainId, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, SignError, _msg)
		return
	}

	txHex, err := txData.MarshalBinary()
	if err != nil {
		_msg := fmt.Sprintf("[ %d ] chain call MarshalBinary error: [ %s ]", msgInfo.ChainId, err.Error())
		logger.Errorf(_msg)
		ReturnError(ctx, SignError, _msg)
		return
	}

	sign := sTypes.Sign{
		Signature: signature,
		TxData:    string(marshalJSON),
		TxHex:     hexutil.Encode(txHex),
	}

	logger.Infof("[Sign Transaction] request ip: [ %s ], chain_id: [ %d ], account: [ %s ], Transaction: [ %s ], signed data: [ %s ]",
		ctx.ClientIP(), msgInfo.ChainId, msgInfo.Account, msgInfo.Transaction, sign.Signature)
	ctx.AbortWithStatusJSON(200, sign)
}

func (s *Service) getMsgData(ctx *gin.Context) ([]byte, ErrCode, error) {
	param := &sTypes.SignRequest{}
	err := ctx.ShouldBind(&param)
	if err != nil {
		return nil, ParamError, err
	}
	return []byte(param.Data), 0, nil
}

func (s *Service) CheckIp(ctx *gin.Context) error {
	err := s.checkIP(ctx.ClientIP())
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) checkIP(ip string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.whitelists[ip]; !ok {
		return ErrIllegalIP
	}
	return nil
}

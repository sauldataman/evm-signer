package yubikey

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/asn1"
	"errors"
	"fmt"
	"github.com/certusone/yubihsm-go"
	"github.com/certusone/yubihsm-go/commands"
	"github.com/certusone/yubihsm-go/connector"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"reflect"
	"strconv"
	"strings"
)

const DefaultURL = "localhost:12345"

var (
	errKey       = errors.New("object keyId MUST > 0")
	errAuthKeyId = errors.New("session auth key MUST exist")
	errPassword  = errors.New("session password MUST exist")
)

type eSig struct {
	R *big.Int
	S *big.Int
}

type HsmClient struct {
	url       string
	keyId     uint16
	authKeyId uint16
	password  string
	sm        *yubihsm.SessionManager
}

func NewHsmClient(url string, keyId, authKeyId uint16, password string) (*HsmClient, error) {

	hsmClient := &HsmClient{url: url, keyId: keyId, authKeyId: authKeyId, password: password}
	if hsmClient.url == "" {
		hsmClient.url = DefaultURL
	}
	// fmt url
	hsmClient.url = fmtUrl(hsmClient.url)
	if err := hsmClient.check(); err != nil {
		return nil, err
	}

	c := connector.NewHTTPConnector(hsmClient.url)
	sm, err := yubihsm.NewSessionManager(c, hsmClient.authKeyId, hsmClient.password)
	if err != nil {
		return nil, fmt.Errorf("new session manager for yubi key error: %s", err)
	}
	hsmClient.sm = sm
	return hsmClient, nil
}

func fmtUrl(url string) string {
	// 如果用户自定义的 url，则需要去掉 http:// 或者 https://
	url = strings.Replace(url, "http://", "", 1)
	url = strings.Replace(url, "https://", "", 1)
	return url
}

func (hc *HsmClient) check() error {
	if hc.keyId < 0 {
		return errKey
	}

	if hc.authKeyId < 0 {
		return errAuthKeyId
	}

	if hc.password == "" {
		return errPassword
	}
	return nil
}

func (hc *HsmClient) GetUrl() string {
	return hc.url
}

func (hc *HsmClient) SetUrl(url string) {
	hc.url = url
}

func (hc *HsmClient) GetPubKeyId() uint {
	return 0
}

func (hc *HsmClient) SetPubKeyId(keyId uint) {}

func (hc *HsmClient) GetPriKeyId() uint {
	return uint(hc.keyId)
}

func (hc *HsmClient) SetPriKeyId(keyId uint) {
	hc.keyId = uint16(keyId)
}

func (hc *HsmClient) GetPublicKey() (publicKey *ecdsa.PublicKey, address common.Address, err error) {
	pubKey, err := hc.getPubkey()
	if err != nil {
		return nil, [20]byte{}, err
	}
	addr := crypto.PubkeyToAddress(*pubKey)
	return pubKey, addr, nil
}

func (hc *HsmClient) getPubkey() (*ecdsa.PublicKey, error) {
	pubKey, err := hc.hsmPubkey()
	if err != nil {
		return nil, err
	}
	hex := "04" + common.Bytes2Hex(pubKey)
	pk, err := crypto.UnmarshalPubkey(common.Hex2Bytes(hex))
	if err != nil {
		return nil, err
	}
	return pk, nil
}

func (hc *HsmClient) hsmPubkey() ([]byte, error) {
	command, err := commands.CreateGetPubKeyCommand(hc.keyId)
	if err != nil {
		return nil, fmt.Errorf("getPubKey error: %s", err)
	}
	res, err := hc.sm.SendEncryptedCommand(command)
	if err != nil {
		return nil, fmt.Errorf("send encrypted via hsm session error: %s", err)
	}
	parsedRes, matched := res.(*commands.GetPubKeyResponse)
	if !matched {
		tp := reflect.TypeOf(res)
		return nil, fmt.Errorf("invalid response type %s", tp)
	}
	return parsedRes.KeyData, nil
}

func (hc *HsmClient) checkHsmKey() error {
	getKeyCommand, err := commands.CreateListObjectsCommand(commands.NewIDOption(hc.keyId))
	if err != nil {
		return err
	}
	getKeyResult, err := hc.sm.SendEncryptedCommand(getKeyCommand)
	if err != nil {
		return err
	}
	parsedres, matched := getKeyResult.(*commands.ListObjectsResponse)
	if !matched {
		return errors.New("invalid response type")
	}
	if len(parsedres.Objects) == 0 {
		return errors.New("key not found")
	}
	return nil
}

// HsmSign 调用hsm签名
func (hc *HsmClient) hsmSign(hashData []byte) ([]byte, error) {
	err := hc.checkHsmKey()
	if err != nil {
		return nil, err
	}
	hsmPubkey, err := hc.hsmPubkey()
	if err != nil {
		return nil, err
	}
	prefix := byte(4)
	hsmPubkey = append([]byte{prefix}, hsmPubkey...)
	//logrus.Info("pubkey from hsm: ", common.Bytes2Hex(hsmPubkey))

	//pk, err := crypto.UnmarshalPubkey(hsmPubkey)
	//if err != nil {
	//	return nil, err
	//}
	//addr := crypto.PubkeyToAddress(*pk)
	//logrus.Info("address from pubkey:", addr)

	for {
		command, err := commands.CreateSignDataEcdsaCommand(hc.keyId, hashData)
		if err != nil {
			return nil, err
		}
		res, err := hc.sm.SendEncryptedCommand(command)
		if err != nil {
			return nil, err
		}
		parsedres, matched := res.(*commands.SignDataEcdsaResponse)
		if !matched {
			return nil, errors.New("invalid response type")
		}
		sig := parsedres.Signature
		var esig eSig
		_, err = asn1.Unmarshal(sig, &esig)
		if err != nil {
			return nil, err
		}
		ethFormatSig := []byte{}
		ethFormatSig = append(ethFormatSig, esig.R.Bytes()...)
		ethFormatSig = append(ethFormatSig, esig.S.Bytes()...)
		//logrus.Info("sig from hsm:", common.Bytes2Hex(sig))
		valid := crypto.VerifySignature(hsmPubkey, hashData, ethFormatSig)
		if !valid {
			continue
		} else {
			// find v
			sigWith0 := append(ethFormatSig, byte(0))
			pubkey, err := crypto.Ecrecover(hashData, sigWith0)
			if err != nil {
				return nil, err
			}
			//logrus.Infof("0 as v,recover: %s", common.Bytes2Hex(pubkey))
			//logrus.Infof("0 as v,hsm    : %s", common.Bytes2Hex(hsmPubkey))
			if bytes.Equal(pubkey, hsmPubkey) {
				return sigWith0, nil
			}
			sigWith1 := append(ethFormatSig, byte(1))
			pubkey, err = crypto.Ecrecover(hashData, sigWith1)
			if err != nil {
				return nil, err
			}
			//logrus.Infof("1 as v,recover: %s", common.Bytes2Hex(pubkey))
			//logrus.Infof("1 as v,hsm    : %s", common.Bytes2Hex(hsmPubkey))
			if bytes.Equal(pubkey, hsmPubkey) {
				return sigWith1, nil
			}
			return nil, errors.New("sign error with 0 | 1 v value")
		}
	}
}

func (hc *HsmClient) SignTx(chainId *big.Int, tx *types.Transaction) ([]byte, *types.Transaction, error) {
	//signer := types.NewLondonSigner(chainId)
	signer := types.LatestSignerForChainID(chainId)
	hash := signer.Hash(tx)
	sig, err := hc.hsmSign(hash[:])
	if err != nil {
		return nil, nil, fmt.Errorf("call hsm sign error: %s", err)
	}
	//logrus.Info("tx with signature: ", common.Bytes2Hex(sig))
	signedTx, err := tx.WithSignature(signer, sig)
	if err != nil {
		return nil, nil, fmt.Errorf("call WithSignature error: %s", err)
	}
	return sig, signedTx, err
}

func (hc *HsmClient) SignEip712(hash []byte) ([]byte, error) {
	return hc.hsmSign(hash)
}

func (hc *HsmClient) SignMessage(msg string) ([]byte, error) {
	ethMessage := append([]byte("\x19Ethereum Signed Message:\n"+strconv.Itoa(len(msg))), msg...)
	hash := crypto.Keccak256(ethMessage)
	return hc.hsmSign(hash[:])
}

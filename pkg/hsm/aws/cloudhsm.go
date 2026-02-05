package aws

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/asn1"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/miekg/pkcs11"
	"log"
	"math/big"
	"strconv"
)

var (
	errPinNull          = errors.New("pin code is null")
	errPbKeyIdEqPvKeyId = errors.New("publicKeyId == privateKeyId")
	errPbKeyIdLt        = errors.New("publicKeyId < 0")
	errPvKeyIdLt        = errors.New("privateKeyId < 0")
)

const LibPath = "/opt/cloudhsm/lib/libcloudhsm_pkcs11.so"

type HsmClient struct {
	p                *pkcs11.Ctx
	session          *pkcs11.SessionHandle
	pbKeyId, pvKeyId int
}

type eSig struct {
	R, S *big.Int
}

type ecdsaSignature eSig

func NewAwsHsm(pin, libPath string, pbKeyId, pvKeyId int) (*HsmClient, error) {
	if libPath == "" {
		libPath = LibPath
	}

	if pin == "" {
		return nil, errPinNull
	}

	if pbKeyId == pvKeyId {
		return nil, errPbKeyIdEqPvKeyId
	}

	if pbKeyId < 0 {
		return nil, errPbKeyIdLt
	}

	if pvKeyId < 0 {
		return nil, errPvKeyIdLt
	}

	ctx, session, err := initPkcs11(pin, libPath)
	if err != nil {
		return nil, err
	}

	return &HsmClient{
		p:       ctx,
		session: session,
		pbKeyId: pbKeyId,
		pvKeyId: pvKeyId,
	}, nil
}

func (ac *HsmClient) GetUrl() string {
	return ""
}

func (ac *HsmClient) SetUrl(url string) {}

func (ac *HsmClient) GetPubKeyId() uint {
	return uint(ac.pbKeyId)
}

func (ac *HsmClient) SetPubKeyId(keyId uint) {
	ac.pbKeyId = int(keyId)
}

func (ac *HsmClient) GetPriKeyId() uint {
	return uint(ac.pvKeyId)
}

func (ac *HsmClient) SetPriKeyId(keyId uint) {
	ac.pvKeyId = int(keyId)
}

func (ac *HsmClient) GetPublicKey() (*ecdsa.PublicKey, common.Address, error) {
	pubKeyBytes, err := ac.p.GetAttributeValue(
		*ac.session,
		pkcs11.ObjectHandle(ac.pbKeyId),
		[]*pkcs11.Attribute{pkcs11.NewAttribute(pkcs11.CKA_EC_POINT, nil)},
	)
	if err != nil {
		return nil, [20]byte{}, fmt.Errorf("get publicKey by publicKey Id error: %s", err)
	}

	if len(pubKeyBytes) == 0 {
		return nil, [20]byte{}, fmt.Errorf("publicKey bytes is null, please check aws cloudhsm keypair")
	}

	if pubKeyBytes[0].Value == nil {
		return nil, [20]byte{}, fmt.Errorf("publicKey bytes is null, please check aws cloudhsm keypair type")
	}

	pubKey, err := parseECDSAPublicKey(pubKeyBytes[0].Value)
	addr := ethCrypto.PubkeyToAddress(*pubKey)

	return pubKey, addr, nil
}

// 解析 ECDSA 公钥
func parseECDSAPublicKey(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	var raw asn1.RawValue

	_, err := asn1.Unmarshal(pubKeyBytes, &raw)
	if err != nil {
		return nil, fmt.Errorf("asns unmarshal error: %s", err.Error())
	}

	uncompressedPubKey := raw.Bytes
	curve := ethCrypto.S256() // 以太坊使用 secp256k1 曲线

	if len(uncompressedPubKey) != 65 || uncompressedPubKey[0] != 0x04 {
		return nil, errors.New("invalid uncompressed public key format")
	}

	pubKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(uncompressedPubKey[1:33]),
		Y:     new(big.Int).SetBytes(uncompressedPubKey[33:]),
	}

	if !curve.IsOnCurve(pubKey.X, pubKey.Y) {
		return nil, errors.New("public key is not on the curve")
	}
	return pubKey, nil
}

func (ac *HsmClient) SignTx(chainId *big.Int, tx *types.Transaction) ([]byte, *types.Transaction, error) {
	//signer := types.NewLondonSigner(chainId)
	signer := types.LatestSignerForChainID(chainId)
	hash := signer.Hash(tx)
	signed, err := ac.hsmSign(hash[:])
	if err != nil {
		return nil, nil, err
	}
	log.Printf("signed: %s \n", signed)
	signedTx, err := tx.WithSignature(signer, signed)
	if err != nil {
		return nil, nil, fmt.Errorf("call WithSignature error: %s", err)
	}
	return signed, signedTx, nil
}

func (ac *HsmClient) SignEip712(hash []byte) ([]byte, error) {
	return ac.hsmSign(hash)
}

func (ac *HsmClient) SignMessage(msg string) ([]byte, error) {
	message := []byte(msg)
	ethMessage := append([]byte("\x19Ethereum Signed Message:\n"+strconv.Itoa(len(message))), message...)
	return ac.hsmSign(ethCrypto.Keccak256(ethMessage))
}

func initPkcs11(pin, lib string) (*pkcs11.Ctx, *pkcs11.SessionHandle, error) {
	p := pkcs11.New(lib)
	if p == nil {
		return nil, nil, fmt.Errorf("failed to load PKCS#11 library")
	}

	// Initialize PKCS#11 library
	err := p.Initialize()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize PKCS#11 library: %s", err)
	}

	// Find the first slot
	slots, err := p.GetSlotList(true)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get slots: %s", err)
	}
	if len(slots) == 0 {
		return nil, nil, fmt.Errorf("no slots found, please check your config")
	}
	slot := slots[0]
	// Open a session
	session, err := p.OpenSession(slot, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open session: %s", err)
	}
	// Login
	err = p.Login(session, pkcs11.CKU_USER, pin)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to login: %s", err)
	}
	return p, &session, nil
}

func (ac *HsmClient) hsmSign(hashData []byte) ([]byte, error) {
	pubkey, addr, err := ac.GetPublicKey()
	if err != nil {
		return nil, err
	}

	log.Printf("addr: %s", addr)
	hsmPubkey := ethCrypto.FromECDSAPub(pubkey)
	for {
		e := ac.p.SignInit(
			*ac.session,
			[]*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_ECDSA, nil)},
			pkcs11.ObjectHandle(ac.pvKeyId),
		)
		if e != nil {
			return nil, fmt.Errorf("failed to sign init: %s", e)
		}

		hsmSig, e := ac.p.Sign(*ac.session, hashData)
		if e != nil {
			return nil, fmt.Errorf("failed to sign: %s", e)
		}

		r := new(big.Int).SetBytes(hsmSig[:len(hsmSig)/2])
		s := new(big.Int).SetBytes(hsmSig[len(hsmSig)/2:])
		parsedSig, e := asn1.Marshal(ecdsaSignature{r, s})
		if e != nil {
			return nil, fmt.Errorf("marshal r, s value for ecdsa signature error: %s", e)
		}

		var esig ecdsaSignature
		_, e = asn1.Unmarshal(parsedSig, &esig)
		if e != nil {
			return nil, fmt.Errorf("unmarshal signature error: %s", e)
		}
		var ethFormatSig []byte
		ethFormatSig = append(ethFormatSig, esig.R.Bytes()...)
		ethFormatSig = append(ethFormatSig, esig.S.Bytes()...)
		valid := ethCrypto.VerifySignature(hsmPubkey, hashData, ethFormatSig)
		if !valid {
			continue
		} else {
			sigWith0 := append(ethFormatSig, byte(0))
			_pubKey, eErr := ethCrypto.Ecrecover(hashData, sigWith0)
			if eErr != nil {
				log.Fatalf("ecrCover0 error: %s", eErr)
				return nil, eErr
			}
			if bytes.Equal(_pubKey, hsmPubkey) {
				return sigWith0, nil
			}
			sigWith1 := append(ethFormatSig, byte(1))
			_pubKey, err = ethCrypto.Ecrecover(hashData, sigWith1)
			if err != nil {
				log.Fatalf("ecrCover1 error: %s", eErr)
				return nil, err
			}
			if bytes.Equal(_pubKey, hsmPubkey) {
				return sigWith1, nil
			}
			return nil, errors.New("sign error with 0 | 1 v value")
		}
	}
}

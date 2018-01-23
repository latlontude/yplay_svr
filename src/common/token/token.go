package token

import (
	"common/constant"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"time"
)

type Token struct {
	Magic  string `json:"magic"`
	Uuid   int64  `json:"uuid"` //设备唯一标识, 在应用卸载后可能会新生产
	Uin    int64  `json:"uin"`  //用户ID
	Type   int    `json:"type"` //类型
	Ts     int    `json:"ts"`   //token生成日期
	Ver    int    `json:"ver"`  //token版本号
	Device string `json:"device"`
	Os     string `json:"os"`
	AppVer string `json:"appVer"`
}

func GeneToken(uin int64, ver int, ttl int, uuid int64, device, os, appVer string) (tokenstr string, err error) {

	b := make([]byte, 12)
	rand.Read(b)

	ts := time.Now().Unix() + int64(ttl)
	typ := 1

	token := Token{string(b), uuid, uin, typ, int(ts), ver, device, os, appVer}

	data, err := json.Marshal(token)
	if err != nil {
		return
	}

	enc, err := AesEncrypt(data, ver)
	if err != nil {
		return
	}

	tokenstr = base64.StdEncoding.EncodeToString(enc)

	return
}

func GetUuidFromTokenString(tokenstr string, ver int) (uuid int64, err error) {

	t, err := DecryptToken(tokenstr, ver)
	if err != nil {
		return
	}

	uuid = t.Uuid

	return
}

func DecryptToken(tokenstr string, ver int) (t Token, err error) {

	data, err := base64.StdEncoding.DecodeString(tokenstr)
	if err != nil {
		return
	}

	org, err := AesDecrypt(data, ver)
	if err != nil {
		return
	}

	err = json.Unmarshal(org, &t)
	if err != nil {
		return
	}

	return
}

func AesEncrypt(src []byte, ver int) (dst []byte, err error) {

	block, err := aes.NewCipher([]byte(constant.TOKEN_YPLAY_AES_KEY))
	if err != nil {
		return
	}

	dst = make([]byte, aes.BlockSize+len(src))

	iv := dst[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)

	stream.XORKeyStream(dst[aes.BlockSize:], src)

	return
}

func AesDecrypt(src []byte, ver int) (dst []byte, err error) {

	if len(src) < aes.BlockSize {
		err = errors.New("ciphertext too short")
		return
	}

	block, err := aes.NewCipher([]byte(constant.TOKEN_YPLAY_AES_KEY))
	if err != nil {
		return
	}

	dst = src[0:]

	iv := dst[:aes.BlockSize]
	dst = dst[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	stream.XORKeyStream(dst, dst)

	return
}

func AesDecryptWxData(src, skey, iv string) (dst []byte, err error) {

	skeyNew, err := base64.StdEncoding.DecodeString(skey)
	if err != nil {
		return
	}

	ivNew, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return
	}

	srcNew, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return
	}

	if len(srcNew) < aes.BlockSize {
		err = errors.New("ciphertext too short")
		return
	}

	if len(srcNew)%aes.BlockSize != 0 {
		err = errors.New("ciphertext is not a multiple of the block size")
		return
	}

	block, err := aes.NewCipher([]byte(skeyNew))
	if err != nil {
		return
	}

	dst = srcNew[0:]

	mode := cipher.NewCBCDecrypter(block, ivNew)

	mode.CryptBlocks(dst, dst)

	dst = PKCS7UnPadding(dst)

	return
}

func PKCS7UnPadding(src []byte) []byte {

	length := len(src)
	unpadding := int(src[length-1])

	return src[:(length - unpadding)]
}

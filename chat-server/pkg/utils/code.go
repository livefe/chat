package utils

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

const nicknameAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

func GenerateCode(length int) string {
	if length <= 0 {
		return ""
	}
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return ""
		}
		code[i] = byte('0' + n.Int64())
	}
	return string(code)
}

func RandomToken(byteLen int) string {
	if byteLen <= 0 {
		return ""
	}
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return GenerateCode(24)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func RandomNickname() string {
	const n = 8
	out := make([]byte, n)
	max := big.NewInt(int64(len(nicknameAlphabet)))
	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "user" + GenerateCode(4)
		}
		out[i] = nicknameAlphabet[v.Int64()]
	}
	return string(out)
}

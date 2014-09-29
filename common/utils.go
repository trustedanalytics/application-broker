package common

import (
	"crypto/rand"
	"encoding/base64"
)

func genPassword(key string, length int) string {
	var bytes = make([]byte, length)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = key[b%byte(len(key))]
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

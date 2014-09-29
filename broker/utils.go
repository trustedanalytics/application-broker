package broker

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	passwordLength   = 24
	passwordEncoding = base64.URLEncoding
)

func getRandomKey() []byte {
	k := make([]byte, passwordLength)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}

func genPassword() (string, error) {
	if key := getRandomKey(); key == nil {
		return "", errors.New("error while generating random key")
	} else {
		return passwordEncoding.EncodeToString(key), nil
	}
}

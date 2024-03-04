package engine_impl

import (
	"Yandex/internal/engine"
	"encoding/base64"
	"hash/crc64"
	"strings"
)

var _ engine.Engine = (*HashImpl)(nil)

type HashImpl struct {
}

func New() *HashImpl {
	return &HashImpl{}
}

func (_ *HashImpl) Get(url string) (string, error) {
	tab := crc64.MakeTable(crc64.ECMA)
	hash := crc64.New(tab)
	writer := strings.Builder{}
	encoder := base64.NewEncoder(base64.StdEncoding, &writer)
	_, err := hash.Write([]byte(url))
	if err != nil {
		return "", err
	}
	result := hash.Sum(nil)
	_, err = encoder.Write(result)
	return writer.String(), err
}

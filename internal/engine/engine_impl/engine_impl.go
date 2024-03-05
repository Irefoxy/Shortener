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
	writer := strings.Builder{}

	tab := crc64.MakeTable(crc64.ECMA)
	hash := crc64.New(tab)
	encoder := base64.NewEncoder(base64.StdEncoding, &writer)
	defer encoder.Close()

	if _, err := hash.Write([]byte(url)); err != nil {
		return "", err
	}
	result := hash.Sum(nil)
	if _, err := encoder.Write(result); err != nil {
		return "", err
	}
	return writer.String(), nil
}

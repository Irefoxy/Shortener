package parser

import (
	"encoding/base64"
	"hash/crc64"
	"strings"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (_ *Parser) Parse(input []byte) (string, error) {
	hash, err := getHash(input)
	if err != nil {
		return "", err
	}
	str, err2 := encodeHashToString(hash)
	if err2 != nil {
		return "", err2
	}
	return str, nil
}

func encodeHashToString(result []byte) (string, error) {
	writer := strings.Builder{}
	encoder := base64.NewEncoder(base64.StdEncoding, &writer)
	defer encoder.Close()
	if _, err := encoder.Write(result); err != nil {
		return "", err
	}
	return writer.String(), nil
}

func getHash(input []byte) ([]byte, error) {
	tab := crc64.MakeTable(crc64.ECMA)
	hash := crc64.New(tab)
	if _, err := hash.Write(input); err != nil {
		return nil, err
	}
	result := hash.Sum(nil)
	return result, nil
}

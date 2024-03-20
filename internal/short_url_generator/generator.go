package short_url_generator

import (
	"encoding/base64"
	"hash/crc64"
	"strings"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

// Parse Collision resolution should be implemented.
func (_ *Parser) Generate(input string) (string, error) {
	hash, err := getHash([]byte(input))
	if err != nil {
		return "", err
	}
	str, err := encodeHashToString(hash)
	if err != nil {
		return "", err
	}
	return str, nil
}

func encodeHashToString(result []byte) (string, error) {
	writer := strings.Builder{}
	encoder := base64.NewEncoder(base64.URLEncoding, &writer)
	if _, err := encoder.Write(result); err != nil {
		return "", err
	}
	encoder.Close()
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

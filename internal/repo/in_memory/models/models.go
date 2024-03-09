package models

import (
	"encoding/json"
	"os"
)

type StorageUnit struct {
	Uuid     int    `json:"uuid"`
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

type FileInfo struct {
	Name    string
	File    *os.File
	Encoder *json.Encoder
}

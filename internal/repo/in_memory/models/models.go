package models

import (
	"encoding/json"
	"os"
)

type FileInfo struct {
	Name    string
	File    *os.File
	Encoder *json.Encoder
}

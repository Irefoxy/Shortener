package in_memory

import (
	"Yandex/internal/repo/in_memory/models"
	"encoding/json"
	"io"
	"os"
	"strings"
)

type Implementation struct {
	data map[string]string
	info models.FileInfo
}

func New(name string) *Implementation {
	return &Implementation{
		data: make(map[string]string),
		info: models.FileInfo{
			Name: name,
		},
	}
}

func (i *Implementation) Init() error {
	f, err := os.OpenFile(i.info.Name, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	i.info.File = f
	data, err := readFile(f)
	if err != nil {
		return err
	}
	i.data = data
	i.info.Encoder = json.NewEncoder(f)
	return nil
}

func readFile(f *os.File) (map[string]string, error) {
	result := make(map[string]string)
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return result, nil
	}
	str := strings.Split(string(bytes), "\n")
	for _, line := range str {
		if strings.TrimSpace(line) == "" {
			break
		}
		var data models.StorageUnit
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			return nil, err
		}
		result[data.Short] = data.Original
	}
	return result, nil
}

func (i *Implementation) Get(hash string) (string, bool) {
	v, ok := i.data[hash]
	return v, ok
}

func (i *Implementation) Set(hash, url string) error {
	if _, ok := i.data[hash]; !ok {
		err := i.info.Encoder.Encode(models.StorageUnit{
			Uuid:     len(i.data) + 1,
			Short:    hash,
			Original: url,
		})
		if err != nil {
			return err
		}
	}
	i.data[hash] = url

	return nil
}

func (i *Implementation) Close() {
	i.info.File.Close()
}

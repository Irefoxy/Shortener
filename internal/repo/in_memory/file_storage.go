package in_memory

import (
	"Yandex/internal/models"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
)

var _ FileStorage[int] = (*JSONFileStorage[int])(nil)

type JSONFileStorage[T any] struct {
	name    string
	file    *os.File
	encoder *json.Encoder
	mu      sync.Mutex
}

func NewJSONFileStorage[T any](name string) *JSONFileStorage[T] {
	return &JSONFileStorage[T]{
		name: name,
	}
}

func (i *JSONFileStorage[T]) Open() (err error) {
	defer func() {
		if err == nil {
			i.encoder = json.NewEncoder(i.file)
		}
	}()
	if i.empty() {
		return models.ErrorFileNameNotGiven
	}
	if i.file != nil {
		return models.ErrorFileAlreadyOpened
	}
	i.file, err = os.OpenFile(i.name, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (i *JSONFileStorage[T]) LoadAll() (str []T, err error) {
	if !i.IsOpened() {
		return nil, models.ErrorFileNotOpened
	}
	data, err := i.readFile()
	if err != nil {
		return nil, err
	}
	return i.convertToStorageUnit(data)
}

func (i *JSONFileStorage[T]) Close() error {
	defer func() {
		i.file = nil
		i.encoder = nil
	}()
	if i.IsOpened() {
		return i.file.Close()
	}
	return nil
}

func (i *JSONFileStorage[T]) Write(a *T) error {
	if !i.IsOpened() {
		return models.ErrorFileNotOpened
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if err := i.encoder.Encode(a); err != nil {
		return err
	}
	return nil
}

func (i *JSONFileStorage[T]) convertToStorageUnit(str []string) ([]T, error) {
	var res []T
	for _, line := range str {
		if strings.TrimSpace(line) == "" {
			break
		}
		var data T
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			return nil, err
		}
		res = append(res, data)
	}
	return res, nil
}

func (i *JSONFileStorage[T]) readFile() ([]string, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	bytes, err := io.ReadAll(i.file)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, nil
	}
	str := strings.Split(string(bytes), "\n")
	return str, nil
}

func (i *JSONFileStorage[T]) empty() bool {
	return i.name == ""
}

func (i *JSONFileStorage[T]) IsOpened() bool {
	return i.file != nil
}

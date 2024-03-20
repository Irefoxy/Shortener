package in_memory

import (
	"encoding/json"
	"os"
)

var _ FileStorage[int] = (*JSONFileStorage[int])(nil)

type JSONFileStorage[T any] struct {
	name string
}

func NewJSONFileStorage[T any](name string) *JSONFileStorage[T] {
	return &JSONFileStorage[T]{
		name: name,
	}
}

func (i *JSONFileStorage[T]) Dump(data []T) (err error) {
	if i.isNotSet() || len(data) == 0 {
		return nil
	}
	file, err := os.OpenFile(i.name, os.O_TRUNC|os.O_CREATE, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	return encoder.Encode(&data)
}

func (i *JSONFileStorage[T]) LoadAll() (str []T, err error) {
	if i.isNotSet() {
		return nil, nil
	}
	data, err := os.ReadFile(i.name)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	err = json.Unmarshal(data, &str)
	if err != nil {
		return nil, err
	}
	return
}

func (i *JSONFileStorage[T]) isNotSet() bool {
	return i.name == ""
}

package in_memory

import (
	"Yandex/internal/models"
	m "Yandex/internal/repo/in_memory/models"
	"Yandex/internal/services/shortener"
	"context"
	"sync"
)

var _ shortener.Repo = (*InMemory)(nil)

//go:generate mockgen -source=in_memory.go -package=mocks -destination=./mocks/mock_filestorage.go
type FileStorage[T any] interface {
	LoadAll() ([]T, error)
	Dump([]T) error
}

type InMemory struct {
	data sync.Map
	file FileStorage[models.Entry]
}

func New(stg FileStorage[models.Entry]) *InMemory {
	return &InMemory{
		file: stg,
	}
}

func (i *InMemory) ConnectStorage() error {
	data, err := i.file.LoadAll()
	if err != nil {
		return err
	}
	i.importData(data)
	return nil
}

func (i *InMemory) Close() error {
	data := i.exportData()
	return i.file.Dump(data)
}

func (i *InMemory) Get(_ context.Context, entry models.Entry) (*models.Entry, error) {
	adapter := m.NewEntryAdapter(entry)
	v, ok := i.data.Load(adapter.Key())
	if !ok {
		return nil, nil
	}
	result := m.KeyValueToEntry(adapter.Key(), v.(m.Value))
	return &result, nil
}

func (i *InMemory) GetAllByUUID(_ context.Context, uuid string) (result []models.Entry, err error) {
	i.data.Range(func(k, v any) bool {
		if k.(m.Key).Id() == uuid {
			result = append(result, m.KeyValueToEntry(k.(m.Key), v.(m.Value)))
		}
		return true
	})
	return
}

func (i *InMemory) Set(_ context.Context, entries []models.Entry) (num int, err error) {
	for _, entry := range entries {
		adapter := m.NewEntryAdapter(entry)
		previous, loaded := i.data.Swap(adapter.Key(), adapter.Value())
		if !loaded || previous.(m.Value).IsDeleted() {
			num++
		}
	}
	return
}

func (i *InMemory) Delete(_ context.Context, entries []models.Entry) error {
	for _, entry := range entries {
		adapter := m.NewEntryAdapter(entry)
		i.data.CompareAndSwap(adapter.Key(), adapter.Value(), adapter.Value().SetDeleted())
	}
	return nil
}

func (i *InMemory) importData(entries []models.Entry) {
	for _, entry := range entries {
		adapter := m.NewEntryAdapter(entry)
		i.data.Store(adapter.Key(), adapter.Value())
	}
}

func (i *InMemory) exportData() (exportedData []models.Entry) {
	i.data.Range(func(key, value any) bool {
		k := key.(m.Key)
		v := value.(m.Value)
		exportedData = append(exportedData, m.KeyValueToEntry(k, v))
		return true
	})
	return exportedData
}

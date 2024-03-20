package in_memory

import (
	"Yandex/internal/models"
	m "Yandex/internal/repo/in_memory/models"
	"context"
	"sync"
)

//var _ shortener.Repo = (*InMemory)(nil)

//go:generate mockgen -source=in_memory.go -package=mocks -destination=./mocks/mock_filestorage.go
type FileStorage[T any] interface {
	LoadAll() ([]T, error)
	Dump([]T) error
}

type InMemory struct {
	data sync.Map
	file FileStorage[models.Entry]
	mu   sync.RWMutex
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
	return i.addDataToMap(data)
}

func (i *InMemory) Close() error {
	data := i.exportData()
	return i.file.Dump(data)
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

func (i *InMemory) Get(_ context.Context, unit models.Entry) (*models.Entry, error) {
	adapter := m.NewEntryAdapter(unit)
	v, ok := i.data.Load(adapter.Key())
	if !ok {
		return nil, nil
	}
	result := m.KeyValueToEntry(adapter.Key(), v.(m.Value))
	return &result, nil
}

func (i *InMemory) GetAllUrlsByUUID(ctx context.Context, uuid string) (result []models.Entry, err error) {
	i.data.Range(func(k, v any) bool {
		if k.(key).id == uuid.Id {
			result = append(result, models.Entry{
				Id:          uuid.Id,
				OriginalUrl: k.(key).original,
				ShortUrl:    v.(string),
			})
		}
		return true
	})
	return
}

func (i *InMemory) Set(ctx context.Context, units []models.Entry) (err error) {
	for _, unit := range units {
		err2 := i.Set(ctx, unit)
		if err2 != nil {
			err = err2
		}
	}
	return
}

func (i *InMemory) Set(_ context.Context, units models.Entry) error {
	_, loaded := i.data.LoadOrStore(key{
		id:       units.Id,
		original: units.OriginalUrl,
	}, units.ShortUrl)
	if loaded {
		return models.ErrorConflict
	}
	if !i.file.IsOpened() {
		return nil
	}
	return i.file.Write(&units)
}

func (i *InMemory) Delete(ctx context.Context, units []models.Entry) error {

}

func (i *InMemory) addDataToMap(units []models.Entry) error {
	for _, unit := range units {
		i.data.Store(key{
			id:       unit.Id,
			original: unit.OriginalUrl,
		}, unit.ShortUrl)
	}
	return nil
}

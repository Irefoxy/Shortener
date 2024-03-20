package in_memory

import (
	"Yandex/internal/api/gin_api"
	"Yandex/internal/models"
	"context"
	"sync"
)

var _ gin_api.Repo = (*InMemory)(nil)

//go:generate mockgen -source=in_memory.go -package=mocks -destination=./mocks/mock_filestorage.go
type FileStorage[T any] interface {
	Open() error
	LoadAll() ([]T, error)
	Close() error
	Write(*T) error
	IsOpened() bool
}

type InMemory struct {
	data sync.Map
	file FileStorage[models.Entry]
	mu   sync.RWMutex
}

type key struct {
	id       string
	original string
}

func New(stg FileStorage[models.Entry]) *InMemory {
	return &InMemory{
		file: stg,
	}
}

func (i *InMemory) ConnectStorage(_ context.Context) error {
	err := i.file.Open()
	if err != nil {
		return err
	}
	data, err := i.file.LoadAll()
	if err != nil {
		return err
	}
	return i.addDataToMap(data)
}

func (i *InMemory) Close(_ context.Context) error {
	return i.file.Close()
}

func (i *InMemory) Get(_ context.Context, units models.Entry) (*models.Entry, error) {
	v, ok := i.data.Load(key{
		id:       units.Id,
		original: units.OriginalUrl,
	})
	if !ok {
		return nil, nil
	}
	units.ShortUrl = v.(string)
	return &units, nil
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

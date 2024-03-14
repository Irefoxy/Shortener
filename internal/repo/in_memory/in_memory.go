package in_memory

import (
	"Yandex/internal/models"
	"Yandex/internal/service/gin_srv"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
)

var _ gin_srv.Repo = (*InMemory)(nil)

type InMemory struct {
	data sync.Map
	info *FileInfo
	mu   sync.RWMutex
}

type FileInfo struct {
	Name    string
	File    *os.File
	Encoder *json.Encoder
	Opened  bool
}

type key struct {
	id       string
	original string
}

func NewFileInfo(name string) *FileInfo {
	return &FileInfo{
		Name: name,
	}
}

func New(name string) *InMemory {
	return &InMemory{
		info: NewFileInfo(name),
	}
}

func (i *InMemory) Init(_ context.Context) error {
	if i.info.empty() {
		return nil
	}
	if err := i.info.openFile(); err != nil {
		return err
	}
	if err := i.loadData(); err != nil {
		return err
	}
	return nil
}

func (i *InMemory) Close(_ context.Context) error {
	return i.info.File.Close()
}

func (i *InMemory) Get(_ context.Context, units models.ServiceUnit) (*models.ServiceUnit, error) {
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

func (i *InMemory) GetAllUrls(_ context.Context, unit models.ServiceUnit) (result []models.ServiceUnit, err error) {
	i.data.Range(func(k, v any) bool {
		if k.(key).id == unit.Id {
			result = append(result, models.ServiceUnit{
				Id:          unit.Id,
				OriginalUrl: k.(key).original,
				ShortUrl:    v.(string),
			})
		}
		return true
	})
	return
}

func (i *InMemory) SetBatch(ctx context.Context, units []models.ServiceUnit) (err error) {
	for _, unit := range units {
		err2 := i.Set(ctx, unit)
		if err2 != nil {
			err = err2
		}
	}
	return
}

func (i *InMemory) Set(_ context.Context, units models.ServiceUnit) error {
	_, loaded := i.data.LoadOrStore(key{
		id:       units.Id,
		original: units.OriginalUrl,
	}, units.ShortUrl)
	if !loaded && i.info.Encoder != nil {
		if err := i.info.Encoder.Encode(&units); err != nil {
			return err
		}
	}
	if loaded {
		return models.ErrorConflict
	}
	return nil
}

func (i *FileInfo) empty() bool {
	return i.Name == ""
}

func (i *FileInfo) openFile() error {
	var err error
	i.File, err = os.OpenFile(i.Name, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	i.Opened = true
	return nil
}

func (i *InMemory) loadData() error {
	err := i.readFile()
	if err != nil {
		errors.Join(models.ErrorFailedToLoadData, err)
	}
	i.info.Encoder = json.NewEncoder(i.info.File)
	return nil
}

func (i *InMemory) readFile() error {
	bytes, err := io.ReadAll(i.info.File)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return nil
	}
	str := strings.Split(string(bytes), "\n")
	err = i.addDataToMap(str)
	if err != nil {
		return err
	}
	return nil
}

func (i *InMemory) addDataToMap(str []string) error {
	for _, line := range str {
		if strings.TrimSpace(line) == "" {
			break
		}
		var data models.ServiceUnit
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			return err
		}
		i.data.Store(key{
			id:       data.Id,
			original: data.OriginalUrl,
		}, data.ShortUrl)
	}
	return nil
}

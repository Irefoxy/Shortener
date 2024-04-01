package models

import "Yandex/internal/models"

type EntryAdapter struct {
	models.Entry
}

func NewEntryAdapter(entry models.Entry) *EntryAdapter {
	return &EntryAdapter{entry}
}

func (a EntryAdapter) Key() Key {
	return Key{
		id:       a.Id,
		original: a.OriginalUrl,
	}
}

func (a EntryAdapter) Value() Value {
	return Value{
		short:   a.ShortUrl,
		deleted: a.DeletedFlag,
	}
}

func KeyValueToEntry(k Key, v Value) models.Entry {
	return models.Entry{
		Id:          k.id,
		OriginalUrl: k.original,
		ShortUrl:    v.short,
		DeletedFlag: v.deleted,
	}
}

func (k Key) Id() string {
	return k.id
}

func (v Value) SetDeleted() Value {
	v.deleted = true
	return v
}

func (v Value) IsDeleted() bool {
	return v.deleted
}

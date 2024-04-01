package models

import "Yandex/internal/models"

type EntryAdapter struct {
	models.Entry
}

func NewEntryAdapter(entry models.Entry) *EntryAdapter {
	return &EntryAdapter{entry}
}

func (a EntryAdapter) Key() Key {
	if a.Id == "" || a.ShortUrl == "" {
		panic("keys fields can't be zero")
	}
	return Key{
		id:    a.Id,
		short: a.ShortUrl,
	}
}

func (a EntryAdapter) Value() Value {
	if a.OriginalUrl == "" {
		panic("value fields can't be zero")
	}
	return Value{
		original: a.OriginalUrl,
		deleted:  a.DeletedFlag,
	}
}

func KeyValueToEntry(k Key, v Value) models.Entry {
	return models.Entry{
		Id:          k.id,
		OriginalUrl: v.original,
		ShortUrl:    k.short,
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

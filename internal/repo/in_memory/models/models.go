package models

type Key struct {
	id    string
	short string
}

type Value struct {
	original string
	deleted  bool
}

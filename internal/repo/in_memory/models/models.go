package models

type Key struct {
	id       string
	original string
}

type Value struct {
	short   string
	deleted bool
}

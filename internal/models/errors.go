package models

type StaticError string

func (e StaticError) Error() string {
	return string(e)
}

const (
	ErrorEmptyBody           = StaticError("request body is empty")
	ErrorConflict            = StaticError("value is already exists")
	ErrorBadContent          = StaticError("wrong content-type")
	ErrorAuthorizationFailed = StaticError("authorization failed")
	ErrorShortURLNotExist    = StaticError("no such short url")
	ErrorDBNotConnected      = StaticError("no db connected")
)

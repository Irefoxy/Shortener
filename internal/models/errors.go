package models

type StaticError string

func (e StaticError) Error() string {
	return string(e)
}

const (
	ErrorDeleted             = StaticError("url is deleted")
	ErrorConflict            = StaticError("value is already exists")
	ErrorNoContent           = StaticError("no content for this user")
	ErrorAuthorizationFailed = StaticError("authorization failed")
	ErrorShortURLNotExist    = StaticError("no such short url")
	ErrorDBNotConnected      = StaticError("no db connected")
	ErrorFileNameNotGiven    = StaticError("no file provided")
	ErrorFileAlreadyOpened   = StaticError("error in loading file")
	ErrorFileNotOpened       = StaticError("file is not opened")
	ErrorContextCanceled     = StaticError("context was cancelled")
	ErrorBadConvertion       = StaticError("can't covert any to necessary type")
	ErrorFailedToStop        = StaticError("failed to stop")
)

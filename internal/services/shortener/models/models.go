package models

import (
	"Yandex/internal/models"
	"context"
	"sync"
)

type Command struct {
	Action       string
	Data         any
	ResponseChan chan<- Response
	Ctx          context.Context
}

type Response struct {
	Err     error
	Entries any
}

type DeleteDispatcher struct {
	Mu       sync.Mutex
	ToDelete [][]models.Entry
}

type BaseContext struct {
	Context   context.Context
	Cancel    context.CancelFunc
	Cancelled chan struct{}
}

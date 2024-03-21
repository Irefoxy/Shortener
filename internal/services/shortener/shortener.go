package shortener

import (
	"Yandex/internal/api/gin_api"
	"Yandex/internal/models"
	m "Yandex/internal/services/shortener/models"
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Generator interface {
	Generate(input string) (string, error)
}

type Repo interface {
	ConnectStorage() error
	Get(ctx context.Context, unit models.Entry) (*models.Entry, error)
	GetAllByUUID(ctx context.Context, uuid string) ([]models.Entry, error)
	Set(ctx context.Context, units []models.Entry) error
	Delete(ctx context.Context, units []models.Entry) error
	Close() error
}

type DbRepo interface {
	Repo
	Ping(ctx context.Context) error
}

var _ gin_api.Service = (*Shortener)(nil)

// Shortener just skeleton should be encapsulated, tested
// idea: handler call operation, all operations are send to a chan
// all operations except deletion run immediately
// deletion fanIN packets and delete
// deletion takes place once n sec and before get or add
type Shortener struct {
	logger      *logrus.Logger
	repo        Repo
	generator   Generator
	requestChan chan m.Command
	wg          sync.WaitGroup
	dispatcher  m.DeleteDispatcher
	context     m.BaseContext
}

func NewShortener(repo Repo, generator Generator, logger *logrus.Logger) *Shortener {
	return &Shortener{
		logger:    logger,
		repo:      repo,
		generator: generator,
	}
}

func (s *Shortener) Stop() error {
	s.context.Cancel()
	select {
	case <-s.context.Cancelled:
		return nil
	case <-time.After(8 * time.Second):
		return models.ErrorFailedToStop
	}
}

func (s *Shortener) Run() error {
	s.context.Context, s.context.Cancel = context.WithCancel(context.Background())
	s.requestChan = make(chan m.Command)
	wg := sync.WaitGroup{}

	go func() {
		<-s.context.Context.Done()
		s.wg.Wait()
		close(s.requestChan)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		for request := range s.requestChan {
			select {
			case <-s.context.Context.Done():
				sendAndClose(request, &m.Response{Err: models.ErrorContextCanceled})
				continue
			case <-request.Ctx.Done():
				continue
			default:
				s.sendResponse(request)
			}
		}
	}()
	go func() {
		defer func() {
			s.context.Cancelled <- struct{}{}
		}()
		for {
			select {
			case <-s.context.Context.Done():
				wg.Wait()
				s.deleteAndLog()
				return
			case <-time.After(30 * time.Second):
				s.deleteAndLog()
			}
		}
	}()

	return nil
}

func (s *Shortener) deleteAndLog() {
	err := s.delete()
	if err != nil {
		s.logger.Warn(err)
	}
}

func (s *Shortener) Add(ctx context.Context, entries []models.Entry) (result []models.Entry, err error) {
	if err := s.checkContext(); err != nil {
		return nil, err
	}
	s.wg.Add(1)
	defer s.wg.Done()
	responseChan := make(chan m.Response, 1)
	s.sendRequest(ctx, entries, "add", responseChan)
	response := <-responseChan
	if response.Entries != nil {
		result, err = convertToType[[]models.Entry](response.Entries)
		if err != nil {
			return nil, err
		}
	}
	return result, response.Err
}

func (s *Shortener) add(ctx context.Context, entries []models.Entry) (result []models.Entry, err error) {
	select {
	case <-ctx.Done():
		return nil, models.ErrorContextCanceled
	default:
		if err = s.generateAndAddShortURLS(entries); err != nil {
			return nil, err
		}
		if err = s.repo.Set(ctx, entries); err != nil && !errors.Is(err, models.ErrorConflict) {
			return nil, err
		}
		return entries, err
	}
}

func (s *Shortener) Ping(ctx context.Context) error {
	if err := s.checkContext(); err != nil {
		return err
	}
	s.wg.Add(1)
	defer s.wg.Done()
	responseChan := make(chan m.Response, 1)
	s.sendRequest(ctx, nil, "ping", responseChan)
	response := <-responseChan
	return response.Err
}

func (s *Shortener) ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return models.ErrorContextCanceled
	default:
		dbRepo, ok := s.repo.(DbRepo)
		if !ok {
			return models.ErrorDBNotConnected
		}
		return dbRepo.Ping(ctx)
	}
}

func (s *Shortener) Get(ctx context.Context, entry models.Entry) (result *models.Entry, err error) {
	if err := s.checkContext(); err != nil {
		return nil, err
	}
	s.wg.Add(1)
	defer s.wg.Done()
	responseChan := make(chan m.Response, 1)
	s.sendRequest(ctx, entry, "get", responseChan)
	response := <-responseChan
	if response.Entries != nil {
		result, err = convertToType[*models.Entry](response.Entries)
		if err != nil {
			return nil, err
		}
	}
	return result, response.Err
}

func (s *Shortener) get(ctx context.Context, entry models.Entry) (*models.Entry, error) {
	select {
	case <-ctx.Done():
		return nil, models.ErrorContextCanceled
	default:
		v, err := s.repo.Get(ctx, entry)
		if err != nil {
			return nil, err
		}
		if v != nil && v.DeletedFlag {
			return nil, models.ErrorDeleted
		}
		return v, nil
	}
}

func (s *Shortener) Delete(ctx context.Context, entries []models.Entry) error {
	if err := s.checkContext(); err != nil {
		return err
	}
	s.wg.Add(1)
	defer s.wg.Done()
	s.sendRequest(ctx, entries, "delete", nil)
	return nil
}

func (s *Shortener) delete() error {
	select { // add timeouts not sure about cancel
	case <-s.context.Context.Done():
		return models.ErrorContextCanceled
	default:
		s.dispatcher.Mu.Lock()
		defer s.dispatcher.Mu.Unlock()
		var entries []models.Entry
		for _, slice := range s.dispatcher.ToDelete {
			for _, entry := range slice {
				entries = append(entries, entry)
			}
		}
		return s.repo.Delete(s.context.Context, entries)
	}
}

func (s *Shortener) GetAll(ctx context.Context, UUID string) (result []models.Entry, err error) {
	if err := s.checkContext(); err != nil {
		return nil, err
	}
	s.wg.Add(1)
	defer s.wg.Done()
	responseChan := make(chan m.Response, 1)
	s.sendRequest(ctx, UUID, "get", responseChan)
	response := <-responseChan
	if response.Entries != nil {
		result, err = convertToType[[]models.Entry](response.Entries)
		if err != nil {
			return nil, err
		}
	}
	return result, response.Err
}

func (s *Shortener) getAll(ctx context.Context, UUID string) ([]models.Entry, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		entries, err := s.repo.GetAllByUUID(ctx, UUID)
		if err != nil {
			return nil, err
		}
		result := excludeDeleted(entries)
		return result, nil
	}
}

func excludeDeleted(entries []models.Entry) []models.Entry {
	newEntries := make([]models.Entry, 0, len(entries))
	for _, entry := range entries {
		if !entry.DeletedFlag {
			newEntries = append(newEntries, entry)
		}
	}
	return newEntries
}

func (s *Shortener) generateAndAddShortURLS(entries []models.Entry) (err error) {
	for _, entry := range entries {
		if entry.OriginalUrl == "" {
			return
		}
		entry.ShortUrl, err = s.generator.Generate(entry.OriginalUrl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Shortener) sendResponse(request m.Command) {
	var response *m.Response
	defer sendAndClose(request, response)
	response = s.provideAction(request)
}

func (s *Shortener) provideAction(request m.Command) *m.Response {
	switch request.Action {
	case "add":
		requests, err := convertToType[[]models.Entry](request.Data)
		if err != nil {
			return &m.Response{Err: err}
		}
		s.deleteAndLog()
		result, err := s.add(request.Ctx, requests)
		return &m.Response{Err: err, Entries: result}
	case "del":
		requests, err := convertToType[[]models.Entry](request.Data)
		if err != nil {
			s.logger.Warn("delete: can;t convert any to []models.Entry")
		}
		s.dispatcher.Mu.Lock()
		defer s.dispatcher.Mu.Unlock()
		s.dispatcher.ToDelete = append(s.dispatcher.ToDelete, requests)
	case "get":
		requests, err := convertToType[models.Entry](request.Data)
		if err != nil {
			return &m.Response{Err: err}
		}
		s.deleteAndLog()
		result, err := s.get(request.Ctx, requests)
		return &m.Response{Err: err, Entries: result}
	case "all":
		requests, err := convertToType[string](request.Data)
		if err != nil {
			return &m.Response{Err: err}
		}
		s.deleteAndLog()
		result, err := s.getAll(request.Ctx, requests)
		return &m.Response{Err: err, Entries: result}
	case "ping":
		err := s.ping(request.Ctx)
		return &m.Response{Err: err}
	}
	return nil
}

func sendAndClose(request m.Command, response *m.Response) {
	if request.ResponseChan != nil {
		if response == nil {
			panic("no such command")
		}
		request.ResponseChan <- *response
		close(request.ResponseChan)
	}
}

func (s *Shortener) sendRequest(ctx context.Context, entries any, action string, responseChan chan<- m.Response) {
	s.requestChan <- m.Command{
		Action:       action,
		Data:         entries,
		ResponseChan: responseChan,
		Ctx:          ctx,
	}
}

func convertToType[T any](data any) (result T, err error) {
	v, ok := data.(T)
	if !ok {
		return result, models.ErrorBadConvertion
	}
	return v, nil
}

func (s *Shortener) checkContext() error {
	select {
	case <-s.context.Context.Done():
		return fmt.Errorf("service: check context: %w", models.ErrorContextCanceled)
	default:
		return nil
	}
}

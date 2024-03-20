package gin_api

import (
	"Yandex/internal/models"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"hash"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const secret = "server" // TODO change to config
const cookieName = "userID"
const parameterName = "id"

type Service interface {
	Run() error
	Stop() error
	Add(ctx context.Context, entries []models.Entry) (result []models.Entry, err error)
	Ping(ctx context.Context) error
	Get(ctx context.Context, entry models.Entry) (*models.Entry, error)
	GetAll(ctx context.Context, UUID string) ([]models.Entry, error)
	Delete(ctx context.Context, entries []models.Entry) error
}

type GinService struct {
	service  Service
	cfg      *models.ApiConf
	logger   *logrus.Logger
	stopChan chan os.Signal
	cookie   *cookieEngine
}

type cookieEngine struct {
	hasher hash.Hash
	equal  func([]byte, []byte) bool
	mu     sync.Mutex
}

func newCookieEngine(secretKey string) *cookieEngine {
	return &cookieEngine{hasher: hmac.New(sha256.New, []byte(secretKey)),
		equal: hmac.Equal}
}

func New(srv Service, cfg *models.ApiConf, logger *logrus.Logger) *GinService {
	return &GinService{srv, cfg, logger, make(chan os.Signal, 1), newCookieEngine(secret)}
}

func (s *GinService) Run() error {
	r := s.init()
	srv := &http.Server{
		Addr:    *s.cfg.HostAddress,
		Handler: r,
	}
	errorChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errorChan <- err
			close(errorChan)
		}
	}()
	signal.Notify(s.stopChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-s.stopChan:
		s.logger.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			s.logger.Warnf("Server Shutdown Failed:%+v", err)
		}
		s.logger.Println("Server gracefully stopped")
	case err := <-errorChan:
		s.logger.Warnf("Server Run error:%+v", err)
	}
	return nil
}

func (s *GinService) Stop() {
	if s.stopChan != nil {
		s.stopChan <- syscall.SIGTERM
		close(s.stopChan)
	}
}

func (s *GinService) init() *gin.Engine {
	r := gin.Default()
	r.Use(s.errorMiddleware, s.authentication, unzipMiddleware, gzip.Gzip(gzip.DefaultCompression))
	r.GET("/*"+parameterName, s.responseLoggerMiddleware, checkAuthentication, s.handleWildcard)
	r.DELETE("/api/user/urls", checkAuthentication, s.handleDelete)
	postGroup := r.Group("/", s.requestLoggerMiddleware, s.setCookie)
	postGroup.POST("/", s.handleUrl)
	postGroup.POST("/shorten", s.handleJsonUrl)
	postGroup.POST("/shorten/batch", s.handleJsonBatch)

	return r
}

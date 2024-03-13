package gin_srv

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
	"syscall"
	"time"
)

const secret = "server" // TODO change to config

type Parser interface {
	Parse(input []byte) (string, error)
}

type Repo interface {
	Init() error
	Get(units models.ServiceUnit) ([]models.ServiceUnit, error)
	Set(units ...models.ServiceUnit) error
	Close() error
}

type DbRepo interface {
	Repo
	Ping() error
}

type GinService struct {
	repo     Repo
	parser   Parser
	cfg      *models.Conf
	logger   *logrus.Logger
	stopChan chan os.Signal
	cookie   *cookieEngine
}

type cookieEngine struct {
	hasher hash.Hash
	equal  func([]byte, []byte) bool
}

func newCookieEngine(secretKey string) *cookieEngine {
	return &cookieEngine{hasher: hmac.New(sha256.New, []byte(secretKey)),
		equal: hmac.Equal}
}

func New(r Repo, e Parser, cfg *models.Conf, logger *logrus.Logger) *GinService {
	return &GinService{r, e, cfg, logger, make(chan os.Signal, 1), newCookieEngine(secret)}
}

func (s *GinService) Run() error {
	if err := s.repo.Init(); err != nil {
		return err
	}
	r := s.init()
	srv := &http.Server{
		Addr:    s.cfg.HostAddress,
		Handler: r,
	}
	errorChan := make(chan error)
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
	s.repo.Close()
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
	r.GET("/*id", s.responseLoggerMiddleware, checkAuthentication, s.handleWildcard)

	postGroup := r.Group("/", s.requestLoggerMiddleware, s.setCookie, checkRequest)
	postGroup.POST("/", s.handleUrl)
	postGroup.POST("/shorten", s.handleJsonUrl)
	postGroup.POST("/shorten/batch", s.handleJsonBatch)

	return r
}

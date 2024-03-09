package service_impl

import (
	"Yandex/internal/engine"
	"Yandex/internal/repo"
	"Yandex/internal/service"
	"context"
	"errors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var _ service.Service = (*ServiceImpl)(nil)

type ServiceImpl struct {
	repo   repo.Repo
	engine engine.Engine
	cfg    *service.Conf
	logger *logrus.Logger
}

func New(r repo.Repo, e engine.Engine, cfg *service.Conf, logger *logrus.Logger) *ServiceImpl {
	return &ServiceImpl{r, e, cfg, logger}
}

func (s *ServiceImpl) Run() error {
	if err := s.repo.Init(); err != nil {
		return err
	}
	r := s.init()

	srv := http.Server{
		Addr:    s.cfg.HostAddress,
		Handler: r,
	}
	errorChan := make(chan error)
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errorChan <- err
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stopChan:
		s.logger.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			s.logger.Warnf("Server Shutdown Failed:%+v", err)
		}
		log.Println("Server gracefully stopped")
	case err := <-errorChan:
		s.logger.Warnf("Server Run error:%+v", err)
	}
	s.repo.Close()
	return nil
}

func (s *ServiceImpl) Stop() {

}

func (s *ServiceImpl) init() *gin.Engine {
	r := gin.Default()
	r.Use(s.errorMiddleware, s.unzipMiddleware, gzip.Gzip(gzip.DefaultCompression))
	r.GET("/*id", s.responseLoggerMiddleware, s.handleRedirect)
	r.POST("/", s.requestLoggerMiddleware, s.checkRequest, s.handleUrl)
	r.POST("/shorten", s.checkRequest, s.handleJsonUrl)

	return r
}

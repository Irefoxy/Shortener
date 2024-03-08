package service_impl

import (
	"Yandex/internal/conf"
	"Yandex/internal/engine"
	"Yandex/internal/repo"
	"Yandex/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var _ service.Service = (*ServiceImpl)(nil)

type ServiceImpl struct {
	repo   repo.Repo
	engine engine.Engine
	cfg    conf.Config
	logger *logrus.Logger
}

func New(r repo.Repo, e engine.Engine, cfg conf.Config, logger *logrus.Logger) *ServiceImpl {
	return &ServiceImpl{r, e, cfg, logger}
}

func (s *ServiceImpl) Run() error {
	r := s.init()
	return r.Run(s.cfg.GetHostAddress())
}

func (s *ServiceImpl) Stop() {

}

func (s *ServiceImpl) init() *gin.Engine {
	r := gin.Default()
	r.Use(s.errorMiddleware)
	r.GET("/*id", s.responseLoggerMiddleware, s.handleRedirect)
	r.POST("/", s.requestLoggerMiddleware, s.checkRequest, s.handleUrl)
	r.POST("/shorten", s.checkRequest, s.handleJsonUrl)

	return r
}

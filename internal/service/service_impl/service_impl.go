package service_impl

import (
	"Yandex/internal/engine"
	"Yandex/internal/repo"
	"Yandex/internal/service"
	"github.com/gin-gonic/gin"
)

var _ service.Service = (*ServiceImpl)(nil)

type ServiceImpl struct {
	repo   repo.Repo
	engine engine.Engine
}

func New(r repo.Repo, e engine.Engine) *ServiceImpl {
	return &ServiceImpl{r, e}
}

func (s *ServiceImpl) Run() error {
	r := s.init()
	return r.Run(":8888")
}

func (s *ServiceImpl) Stop() {

}

func (s *ServiceImpl) init() *gin.Engine {
	r := gin.Default()
	r.Use(s.errorMiddleware)
	r.GET("/*id", s.handleRedirect)
	r.POST("/", s.handleUrl)

	return r
}

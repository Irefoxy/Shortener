package app

import (
	"Yandex/internal/conf"
	"Yandex/internal/parser"
	"Yandex/internal/parser/parser"
	"Yandex/internal/repo/in_memory"
	"Yandex/internal/repo/postgres"
	"Yandex/internal/service/gin_srv"
	"github.com/sirupsen/logrus"
	"os"
)

type Service interface {
	Run() error
	Stop()
}

type Provider struct {
	logger *logrus.Logger
	cfg    *conf.ConfigImpl
	srv    Service
	repo   repo.Repo
	engine engine.Engine
}

func (p *Provider) Service() service.Service {
	if p.srv == nil {
		p.srv = gin_srv.New(p.Repo(), p.Engine(), p.Config().GetServiceConf(), p.Logger())
	}
	return p.srv
}

func (p *Provider) Logger() *logrus.Logger {
	if p.logger == nil {
		p.logger = &logrus.Logger{
			Out:       os.Stdout,
			Formatter: new(logrus.TextFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.InfoLevel,
		}
	}
	return p.logger
}

func (p *Provider) Repo() repo.Repo {
	if p.repo == nil {
		if p.Config().GetDatabaseString() == "" {
			p.repo = in_memory.New(p.Config().GetFileLocation())
		} else {
			p.repo = postgres.New(p.Config().GetDatabaseString())
		}
	}
	return p.repo
}

func (p *Provider) Engine() engine.Engine {
	if p.engine == nil {
		p.engine = parser.New()
	}
	return p.engine
}

func (p *Provider) Config() *conf.ConfigImpl {
	if p.cfg == nil {
		p.cfg = conf.New()
	}
	return p.cfg
}

package app

import (
	"Yandex/internal/api/gin_api"
	"Yandex/internal/conf"
	"Yandex/internal/repo/in_memory"
	"Yandex/internal/repo/postgres"
	"Yandex/internal/short_url_generator"
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
	srv    *gin_api.GinService
	repo   gin_api.Repo
	engine gin_api.Parser
}

func (p *Provider) Service() *gin_api.GinService {
	if p.srv == nil {
		p.srv = gin_api.New(p.Repo(), p.Engine(), p.Config().GetServiceConf(), p.Logger())
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

func (p *Provider) Repo() gin_api.Repo {
	if p.repo == nil {
		if p.Config().GetDatabaseString() == "" {
			p.repo = in_memory.New(p.Config().GetFileLocation())
		} else {
			p.repo = postgres.New(p.Config().GetDatabaseString())
		}
	}
	return p.repo
}

func (p *Provider) Engine() gin_api.Parser {
	if p.engine == nil {
		p.engine = short_url_generator.New()
	}
	return p.engine
}

func (p *Provider) Config() *conf.ConfigImpl {
	if p.cfg == nil {
		p.cfg = conf.New()
		p.cfg.Parse(os.Args[0], os.Args[1:])
	}
	return p.cfg
}

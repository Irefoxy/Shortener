package app

import (
	"Yandex/internal/conf"
	"Yandex/internal/conf/config_impl"
	"Yandex/internal/engine"
	"Yandex/internal/engine/engine_impl"
	"Yandex/internal/repo"
	"Yandex/internal/repo/repo_impl"
	"Yandex/internal/service"
	"Yandex/internal/service/service_impl"
	"github.com/sirupsen/logrus"
	"os"
)

type Provider struct {
	logger *logrus.Logger
	cfg    conf.Config
	srv    service.Service
	repo   repo.Repo
	engine engine.Engine
}

func (p *Provider) Service() service.Service {
	if p.srv == nil {
		p.srv = service_impl.New(p.Repo(), p.Engine(), p.Config(), p.Logger())
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
		p.repo = repo_impl.New()
	}
	return p.repo
}

func (p *Provider) Engine() engine.Engine {
	if p.engine == nil {
		p.engine = engine_impl.New()
	}
	return p.engine
}

func (p *Provider) Config() conf.Config {
	if p.cfg == nil {
		p.cfg = config_impl.New()
	}
	return p.cfg
}

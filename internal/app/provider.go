package app

import (
	"Yandex/internal/api/gin_api"
	"Yandex/internal/conf"
	"Yandex/internal/models"
	"Yandex/internal/repo/in_memory"
	"Yandex/internal/repo/postgres"
	"Yandex/internal/services/shortener"
	"Yandex/internal/short_url_generator"
	"github.com/sirupsen/logrus"
)

type Api interface {
	Run() error
	Stop() error
}

type Provider struct {
	logger    *logrus.Logger
	cfg       *conf.ConfigImpl
	api       Api
	srv       gin_api.Service
	repo      shortener.Repo
	generator shortener.Generator
}

func NewProvider(logger *logrus.Logger, cfg *conf.ConfigImpl) *Provider {
	return &Provider{logger: logger, cfg: cfg}
}

func (p *Provider) Api() Api {
	if p.api == nil {
		p.api = gin_api.New(p.Service(), p.cfg.GetApiConf(), p.logger)
	}
	return p.api
}

func (p *Provider) Service() gin_api.Service {
	if p.srv == nil {
		p.srv = shortener.NewShortener(p.Repo(), p.Generator(), p.logger)
	}
	return p.srv
}

func (p *Provider) Repo() shortener.Repo {
	if p.repo == nil {
		if p.cfg.GetDatabaseString() == "" {
			p.repo = in_memory.New(in_memory.NewJSONFileStorage[models.Entry](p.cfg.GetFileLocation()))
		} else {
			p.repo = postgres.New(p.cfg.GetDatabaseString())
		}
	}
	return p.repo
}

func (p *Provider) Generator() shortener.Generator {
	if p.generator == nil {
		p.generator = short_url_generator.New()
	}
	return p.generator
}

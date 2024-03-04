package app

import (
	"Yandex/internal/engine"
	"Yandex/internal/engine/engine_impl"
	"Yandex/internal/repo"
	"Yandex/internal/repo/repo_impl"
	"Yandex/internal/service"
	"Yandex/internal/service/service_impl"
)

type Provider struct {
	srv    service.Service
	repo   repo.Repo
	engine engine.Engine
}

func (p *Provider) Service() service.Service {
	if p.srv == nil {
		p.srv = service_impl.New(p.Repo(), p.Engine())
	}
	return p.srv
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

package app

import (
	"Yandex/internal/conf"
	"Yandex/internal/repo/in_memory"
	"github.com/sirupsen/logrus"
	"os"
)

type App struct {
	provider *Provider
}

func New() *App {
	cfg := conf.New()
	cfg.Parse(os.Args[0], os.Args[1:])
	logger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
	provider := NewProvider(logger, cfg)
	return &App{provider}
}

func (a App) prepare() error {
	err := a.provider.Repo().ConnectStorage()
	if err != nil {
		repo := a.provider.Repo()
		if _, ok := repo.(*in_memory.InMemory); ok {
			a.provider.logger.Warn("Can't connect file to in memory repo")
		} else {
			return err
		}
	}

	err = a.provider.Service().Run()
	if err != nil {
		return err
	}
	return nil
}

func (a App) close() {
	if err := a.provider.Repo().Close(); err != nil {
		a.provider.logger.Warn("Can't properly close the repo")
	}
	if err := a.provider.Service().Stop(); err != nil {
		a.provider.logger.Warn("Can't properly stop the service")
	}
}

func (a App) Run() error {
	if err := a.prepare(); err != nil {
		return err
	}
	defer a.close()
	return a.provider.Api().Run()
}

func (a App) Stop() error {
	return a.provider.Api().Stop()
}

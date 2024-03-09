package main

import (
	"Yandex/internal/app"
)

func main() {
	provider := app.Provider{}
	srv := provider.Service()
	srv.Run()
}

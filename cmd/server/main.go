package main

import (
	"Yandex/internal/app"
	"log"
)

func main() {
	provider := app.Provider{}
	srv := provider.Service()
	log.Fatal(srv.Run())
}

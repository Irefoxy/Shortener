package main

import (
	"Yandex/internal/app"
	"log"
)

func main() {
	a := app.New()
	if err := a.Run(); err != nil {
		log.Fatalf("Failed to run the app: %v", err)
	}
}

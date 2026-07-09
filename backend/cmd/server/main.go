package main

import (
	"log"

	"tn/backend/internal/app"
	"tn/backend/internal/config"
)

func main() {
	cfg := config.Load()

	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}

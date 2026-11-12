package main

import (
	"log"
	"net/http"

	"brain/internal/app"
	"brain/internal/config"
)

func main() {
	cfg := config.FromEnv()
	if err := cfg.SafeError(); err != nil {
		log.Fatal(err)
	}
	server := app.NewServer()
	log.Fatal(http.ListenAndServe(cfg.HTTPAddress, server.Handler()))
}

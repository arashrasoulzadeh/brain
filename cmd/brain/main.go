package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"

	"brain/internal/app"
	"brain/internal/config"
	"brain/internal/operator"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "bootstrap-operator" {
		bootstrapOperator()
		return
	}
	cfg := config.FromEnv()
	if err := cfg.SafeError(); err != nil {
		log.Fatal(err)
	}
	server := app.NewServer()
	log.Fatal(http.ListenAndServe(cfg.HTTPAddress, server.Handler()))
}

func bootstrapOperator() {
	username := os.Getenv("OPERATOR_USERNAME")
	secret := os.Getenv("OPERATOR_SECRET")
	store := &operator.Store{}
	_, err := store.Bootstrap(username, secret, func(value string) string {
		digest := sha256.Sum256([]byte(value))
		return hex.EncodeToString(digest[:])
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("operator %q bootstrapped", username)
}

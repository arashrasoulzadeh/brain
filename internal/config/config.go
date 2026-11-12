package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	HTTPAddress string
}

func FromEnv() Config {
	return Config{HTTPAddress: os.Getenv("HTTP_ADDRESS")}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.HTTPAddress) == "" {
		return errors.New("HTTP_ADDRESS is required")
	}
	return nil
}

func (c Config) SafeError() error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	return nil
}

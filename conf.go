package main

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	System struct {
		Apipw string `toml:"APIPW"`
	} `toml:"SYSTEM"`
}

func loadConfig(path string) (Config, error) {
	var cfg Config
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

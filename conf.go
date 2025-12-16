package main

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	System struct {
		Apipw string `toml:"APIPW"`
	} `toml:"SYSTEM"`

	Path struct {
		KabuStationExe string `toml:"KABUSTATION_EXE"`
		TradeWebAppURL string `toml:"TRADEWEBAPP_URL"`
	} `toml:"PATH"`
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

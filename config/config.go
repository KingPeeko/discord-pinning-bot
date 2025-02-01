package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var Token string
var BotPrefix string
var config *Config

type Config struct {
	Token     string `json:"token"`
	BotPrefix string `json:"botPrefix"`
}

func readConfig() error {
	file, err := os.ReadFile("./config.json")

	if err != nil {
		fmt.Printf("Could not read file ./config.json")
		return err
	}

	err = json.Unmarshal(file, &config)

	if err != nil {
		fmt.Printf("Failed to unmarshal config.json")
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix

	return nil
}

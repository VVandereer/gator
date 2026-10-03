package config

import (
	"encoding/json"
	"os"
)

const cfgFileName = ".observerconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

func Read() (Config, error) {
	var cfg Config
	path, err := os.UserHomeDir()
	data, err := os.ReadFile(path + "/" + cfgFileName)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func Write(cfg Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	path, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	err = os.WriteFile(path+"/"+cfgFileName, data, 0666)
	return err
}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUsername = username
	return Write(*cfg)
}

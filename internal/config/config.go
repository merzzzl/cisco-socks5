package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-playground/validator/v10"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	CiscoProfile  string `validate:"required" yaml:"cisco_profile"`
	CiscoUsername string `validate:"required" yaml:"cisco_username"`
	CiscoPassword string `validate:"required" yaml:"cisco_password"`
	LocalPassword string `validate:"required" yaml:"local_password"`
}

func LoadConfig() (*Config, error) {
	homedir, _ := os.UserHomeDir()
	file, err := os.ReadFile(homedir + "/" + ".cisco-socks5.yaml")
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}
	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

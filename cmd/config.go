package main

import (
	"errors"
	"flag"
	"os"
	"os/user"

	yaml "gopkg.in/yaml.v2"
)

var errInvalidConfig = errors.New("invalid config of protocols")

type Config struct {
	CiscoUser     string `yaml:"user"`
	CiscoPassword string `yaml:"password"`
	CiscoProfile  string `yaml:"profile"`
	verbose       bool
	debug         bool
	fun           bool
}

func loadConfig() (*Config, error) {
	var cfg Config

	name, _ := os.LookupEnv("SUDO_USER")

	usr, err := user.Lookup(name)
	if err != nil {
		return nil, err
	}

	flag.BoolVar(&cfg.verbose, "verbose", false, "enable verbose logging (default: disabled)")
	flag.BoolVar(&cfg.debug, "debug", false, "enable debug logging (default: disabled)")
	flag.BoolVar(&cfg.fun, "fun", false, "magic!")
	flag.Parse()

	file, err := os.ReadFile(usr.HomeDir + "/" + ".cisco-socks5.yaml")
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	if !cfg.validate() {
		return nil, errInvalidConfig
	}

	return &cfg, nil
}

func (c *Config) validate() bool {
	if c.CiscoUser == "" {
		return false
	}

	if c.CiscoPassword == "" {
		return false
	}

	if c.CiscoProfile == "" {
		return false
	}

	return true
}

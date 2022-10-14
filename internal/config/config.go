package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const configFile = "config/config.yaml"

type Config struct {
	Token    string       `yaml:"token"`
	Storage  StorageConf  `yaml:"storage"`
	Database DatabaseConf `yaml:"database"`
}

type TokenGetter interface {
	GetToken() string
}

type StorageConf struct {
	Mode string `yaml:"mode"`
}

type DatabaseConf struct {
	Dsn      string `yaml:"dsn"`
	MaxTries int    `yaml:"maxTries"`
}

func New() (*Config, error) {
	c := &Config{}

	rawYAML, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}

	err = yaml.Unmarshal(rawYAML, &c)
	if err != nil {
		return nil, errors.Wrap(err, "parsing yaml")
	}

	return c, nil
}

func (c *Config) GetToken() string {
	return c.Token
}

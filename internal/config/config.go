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

type StorageConf struct {
	Mode string `yaml:"mode"`
}

type DatabaseConf struct {
	Dsn      string `yaml:"dsn"`
	MaxTries int    `yaml:"maxTries"`
}

type Service struct {
	config Config
}

func New() (*Service, error) {
	s := &Service{}

	rawYAML, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}

	err = yaml.Unmarshal(rawYAML, &s.config)
	if err != nil {
		return nil, errors.Wrap(err, "parsing yaml")
	}

	return s, nil
}

func (s *Service) Token() string {
	return s.config.Token
}

func (s *Service) Storage() StorageConf {
	return s.config.Storage
}

func (s *Service) Database() DatabaseConf {
	return s.config.Database
}

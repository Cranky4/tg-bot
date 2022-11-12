package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	configFile = "config/config.yaml"
)

type Config struct {
	Token           string            `yaml:"token"`
	Storage         StorageConf       `yaml:"storage"`
	Database        DatabaseConf      `yaml:"database"`
	Logger          LoggerConf        `yaml:"logger"`
	Metrics         MetricsConf       `yaml:"metrics"`
	ReporterMetrics MetricsConf       `yaml:"reporter_metrics"`
	Cache           CacheConf         `yaml:"cache"`
	Redis           RedisConf         `yaml:"redis"`
	MessageBroker   MessageBrokerConf `yaml:"message_broker"`
	GRPC            GRPCConf          `yaml:"grpc"`
	HTTP            HTTPConf          `yaml:"http"`
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

type LoggerConf struct {
	Level string `yaml:"level"`
}

type MetricsConf struct {
	URL  string `yaml:"url"`
	Port int    `yaml:"port"`
}

type CacheConf struct {
	Mode   string `yaml:"mode"`
	Length int    `yaml:"length"`
}

type RedisConf struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type MessageBrokerConf struct {
	Adapter string `yaml:"adapter"`
	Addr    string `yaml:"addr"`
	Queue   string `yaml:"queue"`
	Version string `yaml:"version"`
}

type GRPCConf struct {
	Port int `yaml:"port"`
}
type HTTPConf struct {
	Port int `yaml:"port"`
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

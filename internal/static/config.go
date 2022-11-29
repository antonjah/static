package static

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

const endpointsFileName = "endpoints.yaml"

// Config holds the parsed application configuration
type Config struct {
	Hostname          string `envconfig:"HOSTNAME" default:"127.0.0.1"`
	Port              string `envconfig:"PORT" default:"8080"`
	LogLevel          string `envconfig:"LOG_LEVEL" default:"info"`
	LogPretty         bool   `envconfig:"LOG_PRETTY" default:"false"`
	EndpointsPath     string `envconfig:"ENDPOINTS_PATH" default:""`
	ListenBindAddress string
}

// NewConfig constructs a config from the configuration file and eventual environment
// variables
func NewConfig() Config {
	var config Config

	// read envs
	if err := envconfig.Process("static", &config); err != nil {
		log.Fatal().Err(err).Msg("failed to parse configuration")
	}

	// set bind address
	config.ListenBindAddress = fmt.Sprintf("%s:%s", config.Hostname, config.Port)

	// set endpoints file path
	config.EndpointsPath = filepath.Join(config.EndpointsPath, endpointsFileName)

	// set log level
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to parse log level")
	}
	zerolog.SetGlobalLevel(level)

	// log pretty
	if config.LogPretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return config
}

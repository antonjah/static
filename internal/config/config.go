package config

import (
	"fmt"
	"os"
	"path/filepath"

	env "github.com/caarlos0/env/v11"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const endpointsFileName = "endpoints.yaml"

// Config holds the parsed application configuration
type Config struct {
	Hostname      string `env:"HOSTNAME" envDefault:"127.0.0.1"`
	Port          string `env:"PORT" envDefault:"8080"`
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
	LogPretty     bool   `env:"LOG_PRETTY" envDefault:"false"`
	TLS           bool   `env:"TLS_ENABLED" envDefault:"false"`
	Certificate   string `env:"TLS_CERTIFICATE" envDefault:""`
	Key           string `env:"TLS_KEY" envDefault:""`
	CA            string `env:"TLS_CA" envDefault:""`
	VerifyClient  bool   `env:"TLS_VERIFY_CLIENT" envDefault:"false"`
	EndpointsPath string `env:"ENDPOINTS_PATH" envDefault:""`

	Address       string
	EndpointsFile string
}

// New parses environment variables and returns a Config struct.
func New() Config {
	config, err := env.ParseAs[Config]()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse environment variables")
	}

	config.Address = fmt.Sprintf("%s:%s", config.Hostname, config.Port)

	config.EndpointsFile = filepath.Join(config.EndpointsPath, endpointsFileName)

	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to parse log level")
	}
	zerolog.SetGlobalLevel(level)

	if config.LogPretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return config
}

package config

import (
	"fmt"
	"os"
	"path/filepath"

	env "github.com/caarlos0/env/v11"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const staticapisFileName = "staticapis.yaml"

// Config holds the parsed application configuration
type Config struct {
	Hostname       string `env:"HOSTNAME" envDefault:"127.0.0.1"`
	Port           string `env:"PORT" envDefault:"8080"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"info"`
	LogPretty      bool   `env:"LOG_PRETTY" envDefault:"false"`
	TLS            bool   `env:"TLS_ENABLED" envDefault:"false"`
	Certificate    string `env:"TLS_CERTIFICATE" envDefault:""`
	Key            string `env:"TLS_KEY" envDefault:""`
	CA             string `env:"TLS_CA" envDefault:""`
	VerifyClient   bool   `env:"TLS_VERIFY_CLIENT" envDefault:"false"`
	StaticAPIsPath string `env:"STATICAPIS_PATH" envDefault:""`
	Namespace      string `env:"NAMESPACE" envDefault:""`
	InCluster      bool   `env:"IN_CLUSTER" envDefault:"false"`

	Address        string
	StaticAPIsFile string
}

// New parses environment variables and returns a Config struct.
func New() Config {
	config, err := env.ParseAs[Config]()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse environment variables: %v\n", err)
		os.Exit(1)
	}

	config.Address = fmt.Sprintf("%s:%s", config.Hostname, config.Port)

	// Auto-detect Kubernetes environment if not explicitly set
	if !config.InCluster && config.Namespace == "" {
		// Check if running in Kubernetes by looking for service account
		if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
			config.InCluster = true
		}
	}

	// Get namespace from downward API or service account
	if config.InCluster && config.Namespace == "" {
		// Try to read namespace from downward API
		if nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
			config.Namespace = string(nsBytes)
		} else {
			config.Namespace = "default"
		}
	}

	// Only set StaticAPIsFile if not in cluster mode
	if !config.InCluster {
		config.StaticAPIsFile = filepath.Join(config.StaticAPIsPath, staticapisFileName)
	}

	var zapConfig zap.Config
	if config.LogPretty {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	var level zapcore.Level
	if err := level.Set(config.LogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse log level: %v\n", err)
		os.Exit(1)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	logger, err := zapConfig.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	zap.ReplaceGlobals(logger)

	return config
}

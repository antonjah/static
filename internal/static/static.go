package static

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/antonjah/static/internal/config"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func Run() {
	// read config
	cfg := config.New()

	// read endpoints file
	fh, err := os.Open(cfg.EndpointsFile)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open endpoints file for reading")
	}

	// decode endpoint(s)
	var endpoints Endpoints
	if err = yaml.NewDecoder(fh).Decode(&endpoints); err != nil {
		log.Fatal().Err(err).Msg("failed to open endpoints file for reading")
	}

	mux := http.NewServeMux()

	// append endpoint(s)
	for _, endpoint := range endpoints.Endpoints {
		if err = endpoint.Validate(); err != nil {
			log.Fatal().Str("endpoint", endpoint.Path).Err(err).Msg("validation failed")
		}

		// define a list of supported methods
		endpoint.SetSupported()

		log.Debug().Array("methods", endpoint.SupportedMethods).Msgf("appending endpoint %s", endpoint.Path)

		mux.Handle(endpoint.Path, requestLogger(&endpoint))
	}

	log.Info().Bool("tls", cfg.TLS).Msgf("static is listening on %s", cfg.Address)

	if cfg.TLS {
		caCert, err := os.ReadFile(cfg.CA)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read CA certificate")
		}
		verify := tls.NoClientCert
		if cfg.VerifyClient {
			verify = tls.RequireAndVerifyClientCert
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig := &tls.Config{ClientCAs: caCertPool, ClientAuth: verify}
		server := &http.Server{Addr: cfg.Address, Handler: mux, TLSConfig: tlsConfig}

		if err := server.ListenAndServeTLS(cfg.Certificate, cfg.Key); err != nil {
			log.Fatal().Err(err).Msg("failed to start TLS server")
		}
	} else {
		if err := http.ListenAndServe(cfg.Address, mux); err != nil {
			log.Fatal().Err(err).Msg("failed to start TLS server")
		}
	}

}

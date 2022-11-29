package static

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
)

type defaultHandler struct{}

func (d defaultHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func Run() {
	// read config
	config := NewConfig()

	// read endpoints file
	fh, err := os.Open(config.EndpointsPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open endpoints file for reading")
	}

	// decode endpoint(s)
	var endpoints Endpoints
	if err = yaml.NewDecoder(fh).Decode(&endpoints); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to open endpoints file for reading")
	}

	mux := http.NewServeMux()

	// append endpoint(s)
	for _, endpoint := range endpoints.Endpoints {
		if err = endpoint.Validate(); err != nil {
			log.Fatal().
				Str("endpoint", endpoint.Path).
				Err(err).
				Msg("validation failed")
		}

		// define a list of supported methods
		endpoint.SetSupported()

		log.Debug().
			Array("methods", endpoint.SupportedMethods).
			Msgf("appending endpoint %s", endpoint.Path)

		mux.Handle(endpoint.Path, requestLogger(endpoint))
	}

	mux.Handle("/", requestLogger(defaultHandler{}))

	log.Info().Msgf("static is listening on %s", config.ListenBindAddress)
	if err = http.ListenAndServe(config.ListenBindAddress, mux); err != nil {
		log.Fatal().Err(err).Msg("failed to listen and serve")
	}
}

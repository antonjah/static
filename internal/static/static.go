package static

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/antonjah/static/internal/config"
	staticv1alpha1 "github.com/antonjah/static/pkg/apis/static/v1alpha1"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Server manages the HTTP server and configuration for serving static API responses.
// It supports two modes:
// - Kubernetes mode: watches StaticAPI CRDs directly from the Kubernetes API
// - File mode: watches a configuration file on disk
type Server struct {
	cfg            config.Config
	mux            *http.ServeMux
	mu             sync.RWMutex
	server         *http.Server
	k8sClient      client.Client
	namespace      string
	lastConfigHash string // Track configuration changes
}

// New creates a new Server instance with the given configuration.
// If running in Kubernetes (cfg.InCluster), it initializes the Kubernetes client.
func New(cfg config.Config) *Server {
	server := &Server{
		cfg:       cfg,
		mux:       http.NewServeMux(),
		namespace: cfg.Namespace,
	}

	// Initialize Kubernetes client if in cluster mode
	if cfg.InCluster {
		if err := server.initKubernetesClient(); err != nil {
			zap.L().Fatal("failed to initialize Kubernetes client", zap.Error(err))
		}
	}

	return server
}

// initKubernetesClient initializes the Kubernetes client for in-cluster communication.
func (s *Server) initKubernetesClient() error {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := staticv1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add scheme: %w", err)
	}

	k8sClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	s.k8sClient = k8sClient
	zap.L().Info("Kubernetes client initialized", zap.String("namespace", s.namespace))
	return nil
}

// loadStaticAPIsFromK8s loads StaticAPI resources from Kubernetes and updates the HTTP mux.

func (s *Server) loadStaticAPIsFromK8s(ctx context.Context) error {
	var staticAPIList staticv1alpha1.StaticAPIList
	if err := s.k8sClient.List(ctx, &staticAPIList, client.InNamespace(s.namespace)); err != nil {
		return fmt.Errorf("failed to list StaticAPIs: %w", err)
	}

	// Compute hash of current configuration to detect changes
	configHash, err := computeConfigHash(staticAPIList.Items)
	if err != nil {
		return fmt.Errorf("failed to compute config hash: %w", err)
	}

	// Skip if configuration hasn't changed
	if s.lastConfigHash == configHash {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.mux = http.NewServeMux()

	for _, staticAPIObj := range staticAPIList.Items {
		staticAPI := convertToStaticAPI(staticAPIObj)
		if err := staticAPI.Validate(); err != nil {
			zap.L().Error("validation failed",
				zap.String("name", staticAPIObj.Name),
				zap.String("path", staticAPI.Path),
				zap.Error(err))
			continue
		}

		staticAPI.SetSupported()
		zap.L().Debug("loaded path",
			zap.String("name", staticAPIObj.Name),
			zap.String("path", staticAPI.Path),
			zap.Any("methods", staticAPI.SupportedMethods))
		s.mux.Handle(staticAPI.Path, requestLogger(&staticAPI))
	}

	s.lastConfigHash = configHash
	zap.L().Info("configuration loaded from Kubernetes", zap.Int("apis", len(staticAPIList.Items)))
	return nil
}

// computeConfigHash computes a SHA256 hash of the StaticAPI configuration.
func computeConfigHash(items []staticv1alpha1.StaticAPI) (string, error) {
	data, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// convertToStaticAPI converts a Kubernetes StaticAPI CRD to an internal StaticAPI struct.

func convertToStaticAPI(obj staticv1alpha1.StaticAPI) StaticAPI {
	methods := make([]MethodConfig, len(obj.Spec.Methods))
	for i, m := range obj.Spec.Methods {
		methods[i] = MethodConfig{
			Method:     m.Method,
			StatusCode: m.StatusCode,
			Body:       m.Body,
			Headers:    m.Headers,
		}
	}
	return StaticAPI{
		Path:    obj.Spec.Path,
		Methods: methods,
	}
}

// loadStaticAPIsFromFile loads StaticAPI configuration from a YAML file.

func (s *Server) loadStaticAPIsFromFile() error {
	fh, err := os.Open(s.cfg.StaticAPIsFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	var staticAPIs StaticAPIs
	if err = yaml.NewDecoder(fh).Decode(&staticAPIs); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.mux = http.NewServeMux()

	for _, staticAPI := range staticAPIs.StaticAPIs {
		if err = staticAPI.Validate(); err != nil {
			zap.L().Error("validation failed", zap.String("path", staticAPI.Path), zap.Error(err))
			continue
		}

		staticAPI.SetSupported()
		zap.L().Debug("loaded path", zap.String("path", staticAPI.Path), zap.Any("methods", staticAPI.SupportedMethods))
		s.mux.Handle(staticAPI.Path, requestLogger(&staticAPI))
	}

	zap.L().Info("configuration reloaded from file")
	return nil
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	mux := s.mux
	s.mu.RUnlock()
	mux.ServeHTTP(w, r)
}

// watchKubernetesAPIs polls for StaticAPI changes in Kubernetes every 5 seconds.
func (s *Server) watchKubernetesAPIs(ctx context.Context) {
	// Initial load
	if err := s.loadStaticAPIsFromK8s(ctx); err != nil {
		zap.L().Error("failed to load initial configuration", zap.Error(err))
	}

	// Poll for changes every 5 seconds
	// A more sophisticated implementation would use informers, but this is simpler
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	zap.L().Info("polling for StaticAPI changes", zap.String("namespace", s.namespace))

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.loadStaticAPIsFromK8s(ctx); err != nil {
				zap.L().Error("failed to reload configuration", zap.Error(err))
			}
		}
	}
}

// watchConfigFile watches the configuration file for changes and reloads when detected.
func (s *Server) watchConfigFile(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zap.L().Fatal("failed to create file watcher", zap.Error(err))
	}
	defer watcher.Close()

	configDir := filepath.Dir(s.cfg.StaticAPIsFile)
	if err := watcher.Add(configDir); err != nil {
		zap.L().Fatal("failed to watch configuration directory", zap.String("dir", configDir), zap.Error(err))
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastModTime time.Time
	if info, err := os.Stat(s.cfg.StaticAPIsFile); err == nil {
		lastModTime = info.ModTime()
	}

	zap.L().Info("watching configuration file", zap.String("file", s.cfg.StaticAPIsFile))

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Chmod != 0 {
				continue
			}

			eventBase := filepath.Base(event.Name)
			configBase := filepath.Base(s.cfg.StaticAPIsFile)
			if eventBase == configBase || eventBase == "..data" {
				zap.L().Info("configuration file changed, reloading",
					zap.String("file", event.Name),
					zap.String("op", event.Op.String()))
				time.Sleep(100 * time.Millisecond)
				if err := s.loadStaticAPIsFromFile(); err != nil {
					zap.L().Error("failed to reload configuration", zap.Error(err))
				} else {
					if info, err := os.Stat(s.cfg.StaticAPIsFile); err == nil {
						lastModTime = info.ModTime()
					}
				}
			}
		case <-ticker.C:
			if info, err := os.Stat(s.cfg.StaticAPIsFile); err == nil {
				if info.ModTime().After(lastModTime) {
					lastModTime = info.ModTime()
					zap.L().Info("configuration file changed (poll), reloading")
					if err := s.loadStaticAPIsFromFile(); err != nil {
						zap.L().Error("failed to reload configuration", zap.Error(err))
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			zap.L().Error("watcher error", zap.Error(err))
		}
	}
}

// Run starts the static HTTP server.
// It loads configuration from either Kubernetes or a file based on cfg.InCluster,
// starts watching for changes, and serves HTTP requests.
func Run() {
	cfg := config.New()

	s := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load initial configuration
	if cfg.InCluster {
		if err := s.loadStaticAPIsFromK8s(ctx); err != nil {
			zap.L().Fatal("failed to load initial configuration", zap.Error(err))
		}
		go s.watchKubernetesAPIs(ctx)
	} else {
		if err := s.loadStaticAPIsFromFile(); err != nil {
			zap.L().Fatal("failed to load configuration", zap.Error(err))
		}
		go s.watchConfigFile(ctx)
	}

	s.server = &http.Server{
		Addr:    cfg.Address,
		Handler: s,
	}

	if cfg.TLS {
		verify := tls.NoClientCert
		tlsConfig := &tls.Config{}

		if cfg.VerifyClient {
			verify = tls.RequireAndVerifyClientCert
		}

		if cfg.CA != "" {
			caCert, err := os.ReadFile(cfg.CA)
			if err != nil {
				zap.L().Fatal("failed to read CA certificate", zap.Error(err))
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.ClientCAs = caCertPool
		}

		tlsConfig.ClientAuth = verify
		s.server.TLSConfig = tlsConfig
	}

	go func() {
		zap.L().Info("static is listening",
			zap.Bool("tls", cfg.TLS),
			zap.Bool("in-cluster", cfg.InCluster),
			zap.String("address", cfg.Address))
		var err error
		if cfg.TLS {
			err = s.server.ListenAndServeTLS(cfg.Certificate, cfg.Key)
		} else {
			err = s.server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("server error", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	zap.L().Info("shutting down server")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("server shutdown error", zap.Error(err))
	}
}

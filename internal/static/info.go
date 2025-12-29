package static

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// handleInfo returns JSON information about all configured endpoints.
func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type EndpointInfo struct {
		Path    string   `json:"path"`
		Methods []string `json:"methods"`
	}

	type InfoResponse struct {
		Endpoints []EndpointInfo `json:"endpoints"`
		Total     int            `json:"total"`
	}

	var endpoints []EndpointInfo
	for _, endpoint := range s.endpoints {
		endpoints = append(endpoints, EndpointInfo{
			Path:    endpoint.Path,
			Methods: endpoint.SupportedMethods,
		})
	}

	response := InfoResponse{
		Endpoints: endpoints,
		Total:     len(endpoints),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		zap.L().Error("failed to encode info response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) healthRouter(r chi.Router) {
	r.Route("/health", func(r chi.Router) {
		r.Get("/", s.healthHandler)
	})
}

type HealthCheckInfo struct {
	Key         string        `json:"key" example:"Database"`
	Status      string        `json:"status" example:"Healthy"`
	Description string        `json:"description" example:"Pinged DB"`
	Duration    *JsonDuration `json:"duration" swaggertype:"string" example:"00:00:01.123456"`
}

type HealthCheckResponse struct {
	Status   string            `json:"status" example:"Healthy"`
	Duration *JsonDuration     `json:"duration" swaggertype:"string" example:"00:00:01.123456"`
	Info     []HealthCheckInfo `json:"info"`
}

// Health
// @Summary Healthcheck
// @Description Determine health of the API
// @Tags health
// @Produce json
// @Success 200 {object} server.HealthCheckResponse
// @Router /v1/health [get]
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	dbStart := time.Now()
	dbHealthy, dbMessage := s.db.Health()

	dbHealthInfo := HealthCheckInfo{
		Key:         "Database",
		Status:      getHealthStatus(dbHealthy, true),
		Description: dbMessage,
		Duration:    &JsonDuration{time.Now().Sub(dbStart)},
	}

	var info []HealthCheckInfo
	response := HealthCheckResponse{
		Status:   getHealthStatus(dbHealthy, true),
		Duration: &JsonDuration{time.Now().Sub(start)},
		Info:     append(info, dbHealthInfo),
	}

	s.jsonResponse(w, response)
}

func getHealthStatus(requiredDeps bool, optionalDeps bool) string {
	if !requiredDeps {
		return "Unhealthy"
	}

	if !optionalDeps {
		return "Degraded"
	}

	return "Healthy"
}

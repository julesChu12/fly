package observability

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// HealthCheckServer manages health check and metrics endpoints
type HealthCheckServer struct {
	db   *gorm.DB
	port int
}

// NewHealthCheckServer creates a new health check server
func NewHealthCheckServer(db *gorm.DB, port int) *HealthCheckServer {
	return &HealthCheckServer{
		db:   db,
		port: port,
	}
}

// Start starts the health check server
func (s *HealthCheckServer) Start() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.healthCheckHandler)
	mux.HandleFunc("/health/liveness", s.livenessHandler)
	mux.HandleFunc("/health/readiness", s.readinessHandler)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting health check server on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return server.ListenAndServe()
}

// healthCheckHandler provides overall health status
func (s *HealthCheckServer) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	health := s.checkHealth()

	w.Header().Set("Content-Type", "application/json")
	if health["status"] == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	fmt.Fprintf(w, `{"status":"%s","timestamp":"%s","checks":%v}`,
		health["status"],
		health["timestamp"],
		health["checks"])
}

// livenessHandler checks if the service is alive
func (s *HealthCheckServer) livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"alive"}`)
}

// readinessHandler checks if the service is ready to accept traffic
func (s *HealthCheckServer) readinessHandler(w http.ResponseWriter, r *http.Request) {
	health := s.checkHealth()

	w.Header().Set("Content-Type", "application/json")
	if health["status"] == "healthy" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ready"}`)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"status":"not_ready"}`)
	}
}

// checkHealth performs health checks on dependencies
func (s *HealthCheckServer) checkHealth() map[string]interface{} {
	checks := make(map[string]string)
	status := "healthy"

	// Check database connection
	sqlDB, err := s.db.DB()
	if err != nil {
		checks["database"] = "unhealthy"
		status = "unhealthy"
	} else {
		if err := sqlDB.Ping(); err != nil {
			checks["database"] = "unhealthy"
			status = "unhealthy"
		} else {
			checks["database"] = "healthy"
		}
	}

	return map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
		"checks":    checks,
	}
}

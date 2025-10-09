package observability

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// MetricsServer manages Prometheus metrics endpoint
type MetricsServer struct {
	port int
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port int) *MetricsServer {
	return &MetricsServer{
		port: port,
	}
}

// Start starts the metrics server
func (s *MetricsServer) Start() error {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.HandleFunc("/metrics", s.metricsHandler)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting metrics server on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return server.ListenAndServe()
}

// metricsHandler provides Prometheus-format metrics
func (s *MetricsServer) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)

	// Basic metrics - in production, use prometheus client library
	metrics := `# HELP plutus_up Service up status
# TYPE plutus_up gauge
plutus_up 1

# HELP plutus_build_info Build information
# TYPE plutus_build_info gauge
plutus_build_info{version="1.0.0",service="plutus"} 1

# HELP plutus_transactions_total Total number of transactions processed
# TYPE plutus_transactions_total counter
plutus_transactions_total{type="recharge"} 0
plutus_transactions_total{type="consume"} 0
plutus_transactions_total{type="refund"} 0

# HELP plutus_wallet_balance_total Total wallet balance across all wallets
# TYPE plutus_wallet_balance_total gauge
plutus_wallet_balance_total 0

# HELP plutus_http_requests_total Total HTTP requests
# TYPE plutus_http_requests_total counter
plutus_http_requests_total{method="GET",path="/api/wallets"} 0
plutus_http_requests_total{method="POST",path="/api/transactions/recharge"} 0
`

	fmt.Fprint(w, metrics)
}

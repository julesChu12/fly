package main

import (
	"flag"

	"github.com/julesChu12/fly/mora/adapters/gozero"
	"github.com/julesChu12/fly/mora/pkg/logger"
	"github.com/julesChu12/fly/mora/pkg/observability"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/config"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/handler"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/mora-api.yaml", "the config file")

func main() {
	flag.Parse()

	// Initialize observability
	cfg := observability.Config{
		ServiceName:  "gozero-starter",
		ExporterURL:  "http://localhost:4317", // OTLP endpoint
		SampleRatio:  1.0,
		Environment:  "development",
		ExporterType: "stdout", // Use stdout for demo
	}
	cleanup, err := observability.Init(cfg)
	if err != nil {
		logger.Fatalf("failed to initialize observability: %v", err)
	}
	defer cleanup()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// Configure auth middleware
	authConfig := gozero.AuthMiddlewareConfig{
		Secret:    c.JWT.Secret,
		SkipPaths: []string{"/health", "/login"},
	}

	// Apply auth middleware to protected routes only
	authMiddleware := gozero.AuthMiddleware(authConfig)

	// Public routes (no authentication required)
	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/health",
		Handler: handler.HealthHandler(ctx),
	})

	server.AddRoute(rest.Route{
		Method:  "POST",
		Path:    "/login",
		Handler: handler.LoginHandler(ctx),
	})

	// Protected routes (authentication required)
	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/profile",
		Handler: authMiddleware(handler.ProfileHandler(ctx)),
	})

	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/protected",
		Handler: authMiddleware(handler.ProtectedHandler(ctx)),
	})

	// Business API routes
	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/api/v1/orders",
		Handler: authMiddleware(handler.GetOrdersHandler(ctx)),
	})

	server.AddRoute(rest.Route{
		Method:  "POST",
		Path:    "/api/v1/orders",
		Handler: authMiddleware(handler.CreateOrderHandler(ctx)),
	})

	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/api/v1/users",
		Handler: authMiddleware(handler.GetUsersHandler(ctx)),
	})

	logger.Infof("Starting Go-Zero server with observability at %s:%d", c.Host, c.Port)
	server.Start()
}

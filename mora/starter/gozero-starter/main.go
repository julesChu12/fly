package main

import (
	"flag"
	"fmt"

	"github.com/julesChu12/fly/mora/adapters/gozero"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/config"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/handler"
	"github.com/julesChu12/fly/mora/starter/gozero-starter/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/mora-api.yaml", "the config file")

func main() {
	flag.Parse()

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

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

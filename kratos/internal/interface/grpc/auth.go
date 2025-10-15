package grpc

import (
	"github.com/julesChu12/fly/kratos/internal/infrastructure/custos"
	grpcmiddleware "github.com/julesChu12/fly/kratos/internal/interface/grpc/middleware"
)

// NewGRPCAuthInterceptor creates a new gRPC auth interceptor
func NewGRPCAuthInterceptor(custosClient *custos.Client) *grpcmiddleware.GRPCAuthInterceptor {
	return grpcmiddleware.NewGRPCAuthInterceptor(custosClient)
}

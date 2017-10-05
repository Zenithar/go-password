package server

import (
	"context"

	pb "go.zenithar.org/password/protocol/password"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// -----------------------------------------------------------------------------

func prepareGRPC(context context.Context) (*grpc.Server, error) {
	// gRPC Server settings
	var sopts []grpc.ServerOption

	// gRPC middlewares
	sopts = append(sopts, grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(
				grpc_recovery.WithRecoveryHandler(recoveryFunc)),
			grpc_logrus.StreamServerInterceptor(logrusEntry),
		)),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_opentracing.UnaryServerInterceptor(
					grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
				),
				grpc_prometheus.UnaryServerInterceptor,
				grpc_recovery.UnaryServerInterceptor(
					grpc_recovery.WithRecoveryHandler(recoveryFunc)),
				grpc_logrus.UnaryServerInterceptor(logrusEntry),
			),
		))
	s := grpc.NewServer(sopts...)

	// Password service
	pb.RegisterPasswordServer(s, newServer())

	// Prometheus
	grpc_prometheus.Register(s)

	// Health
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("Password", healthpb.HealthCheckResponse_SERVING)

	// Reflection
	reflection.Register(s)

	return s, nil
}

func init() {
	grpc_logrus.ReplaceGrpcLogger(logrusEntry)
}

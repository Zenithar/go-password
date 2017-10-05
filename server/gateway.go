package server

import (
	"context"
	"net"
	"net/http"
	"time"

	pb "go.zenithar.org/password/protocol/password"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func prepareGateway(ctx context.Context, socketPath string) (http.Handler, error) {
	// gRPC dialup options
	opts := []grpc.DialOption{
		grpc.WithTimeout(10 * time.Second),
		grpc.WithBlock(),
		grpc.WithInsecure(), // Unix socket are insecure by default
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}),
	}

	// gRPC dialup options
	conn, err := grpc.DialContext(ctx, "", opts...)
	if err != nil {
		logrus.WithError(err).Error("fail to dial")
		return nil, err
	}

	// changes json serializer to include empty fields with default values
	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
		runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),
	)

	// Register Gateway endpoints
	err = pb.RegisterPasswordHandler(ctx, gwMux, conn)
	if err != nil {
		return nil, err
	}

	return gwMux, nil
}

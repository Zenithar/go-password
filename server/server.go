package server

import (
	"context"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

var logrusEntry = logrus.NewEntry(logrus.New())

// MicroServer represents a microservice server instance
type MicroServer struct {
	serverName string
	lis        net.Listener
	httpServer *http.Server
	grpcServer *grpc.Server
}

// New returns a microserver instance
func New(serverName string, l net.Listener) *MicroServer {
	return &MicroServer{
		serverName: serverName,
		lis:        l,
	}
}

// -----------------------------------------------------------------------------

// Start the microserver
func (ms *MicroServer) Start() error {
	var err error

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize unix socket
	unixL, err := net.Listen("unix", "service.sock")
	if err != nil {
		logrus.WithError(err).Error("Unable to create unix socket")
		return err
	}

	// tcpMuxer
	tcpMux := cmux.New(ms.lis)

	// Connection dispatcher rules
	grpcL := tcpMux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := tcpMux.Match(cmux.HTTP1Fast())

	// initialize gRPC server instance
	ms.grpcServer, err = prepareGRPC(ctx)
	if err != nil {
		logrus.WithError(err).Error("Unable to initialize gRPC server instance")
		return err
	}

	// initialize HTTP server
	ms.httpServer, err = prepareHTTP(ctx, ms.serverName)
	if err != nil {
		logrus.WithError(err).Error("Unable to initialize HTTP server instance")
		return err
	}

	// Start servers
	go func() {
		if err := ms.grpcServer.Serve(grpcL); err != nil {
			logrus.WithError(err).Fatalln("Unable to start external gRPC server")
		}
	}()
	go func() {
		if err := ms.grpcServer.Serve(unixL); err != nil {
			logrus.WithError(err).Fatalln("Unable to start internal gRPC server")
		}
	}()
	go func() {
		if err := ms.httpServer.Serve(httpL); err != nil {
			logrus.WithError(err).Fatalln("Unable to start HTTP server")
		}
	}()

	return tcpMux.Serve()
}

// Shutdown the microserver
func (ms *MicroServer) Shutdown(ctx context.Context, stopped chan struct{}) {
	ms.grpcServer.GracefulStop()
	ms.httpServer.Shutdown(ctx)
	stopped <- struct{}{}
}

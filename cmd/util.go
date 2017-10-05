package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// mustLoadX509KeyPair handles certificate loading from file
func mustLoadX509KeyPair(certFile, keyFile string) tls.Certificate {
	pair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"certFile": certFile,
			"keyFile":  keyFile,
		}).Fatalln("Unable to load certificates")
	}

	return pair
}

// handleShutdown handles the server shut down error.
func handleShutdown(err error) {
	if err != nil {
		if opErr, ok := err.(*net.OpError); !ok || (ok && opErr.Op != "accept") {
			logrus.WithError(err).Fatal("Error while shutting down server")
		}
	}
}

func grpcClientConnection(ctx context.Context, server string) *grpc.ClientConn {

	// Load Certificates
	certBytes, err := ioutil.ReadFile("./certs/server.ecdsa.crt")
	if err != nil {
		logrus.WithError(err).Fatalln("Unable to read certificate")
	}

	// Create Pool
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certBytes)

	// Client certificate settings
	dcreds := credentials.NewClientTLSFromCert(certPool, server)

	// gRPC dialup options
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(dcreds))
	opts = append(opts, grpc.WithTimeout(10*time.Second))
	opts = append(opts, grpc.WithBlock())

	// Initialize connection
	conn, err := grpc.DialContext(ctx, server, opts...)
	if err != nil {
		logrus.WithError(ctx.Err()).Fatalf("fail to dial: %v", err)
	}

	return conn
}

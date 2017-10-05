package cmd

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.zenithar.org/password/server"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "Launches the server on http://localhost:5555",
	RunE:  serve,
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

func serve(cmd *cobra.Command, args []string) error {

	// Signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	errCh := make(chan error, 1)

	// Initialize listener
	conn, err := net.Listen("tcp", ":5555")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Initialize TLS listener
	config := &tls.Config{
		Certificates: []tls.Certificate{
			mustLoadX509KeyPair("./certs/server.ecdsa.crt", "./certs/server.ecdsa.key"),
			mustLoadX509KeyPair("./certs/server.rsa.crt", "./certs/server.rsa.key"),
		},
		Rand:       rand.Reader,
		NextProtos: []string{},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP384,
			tls.X25519,
		},
	}

	tlsL := tls.NewListener(conn, config)

	// Instanciate the server
	s := server.New("localhost:5555", tlsL)

	// Server
	go func() {
		if err := func() error {
			// Start server
			if err := s.Start(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		logrus.WithError(err).Error("Server failed to start")
		os.Exit(1)
	case sig := <-signalCh:
		logrus.Infof("Signal received '%s'", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stopped := make(chan struct{}, 1)
	go s.Shutdown(ctx, stopped)
	select {
	case <-ctx.Done():
		logrus.Warn("time limit reached, initiating hard shutdown")
		return errors.New("Server is failed")
	case <-stopped:
		logrus.Info("server shutdown completed")
		break
	}
	return nil
}

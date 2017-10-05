package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pb "go.zenithar.org/password/protocol/password"
	"go.zenithar.org/password/version"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.zenithar.org/common/web/utils"
)

var (
	httpServer *http.Server
)

func prepareHTTP(ctx context.Context, serverName string) (*http.Server, error) {
	// Assign a HTTP router
	router := http.NewServeMux()

	// Swagger
	router.HandleFunc("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(w, strings.NewReader(pb.Swagger))
	})

	// Metrics endpoint
	router.Handle("/metrics", prometheus.Handler())

	// Health monitoring endpoint
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
			"status":    "OK",
			"timestamp": time.Now().UTC().Unix(),
		})
	})

	// Service discovery
	router.HandleFunc("/.well-known/finger", func(w http.ResponseWriter, r *http.Request) {
		utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
			"service-name":        "Password",
			"service-description": "Remote password hasher",
			"version":             version.Version,
			"version-full":        fmt.Sprintf("%s (%s-%s)", version.Version, version.Revision, version.Branch),
			"revision":            version.Revision,
			"branch":              version.Branch,
			"build_date":          version.BuildDate,
			"swagger_doc_url":     "/swagger.json",
			"healthz_url":         "/healthz",
			"metric_url":          "/metrics",
			"endpoints":           []string{"grpc", "http"},
		})
	})

	// initialize grpc-gateway
	gw, err := prepareGateway(ctx, "service.sock")
	if err != nil {
		logrus.WithError(err).Error("Unable to initialize gRPC Gateway")
		return nil, err
	}
	router.Handle("/", gw)

	// Return HTTP Server instance
	return &http.Server{
		Addr:         serverName,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}, nil
}

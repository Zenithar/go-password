package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/cmux"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	panichandler "github.com/kazegusuri/grpc-panic-handler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"go.zenithar.org/butcher"
	"go.zenithar.org/common/web/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "go.zenithar.org/password/protocol/password"
	"go.zenithar.org/password/version"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "Launches the server on https://localhost:5555",
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

// -----------------------------------------------------------------------------

type myService struct {
	butch *butcher.Butcher
}

func (m *myService) Encode(c context.Context, s *pb.PasswordReq) (*pb.EncodedPasswordRes, error) {
	res := &pb.EncodedPasswordRes{}

	// Check mandatory fields
	if len(strings.TrimSpace(s.Password)) == 0 {
		res.Error = &pb.Error{
			Code:    http.StatusPreconditionFailed,
			Message: "Password value is mandatory !",
		}
		return res, nil
	}

	// Hash given password
	passwd, err := m.butch.Hash([]byte(s.Password))
	if err != nil {
		res.Error = &pb.Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		return res, nil
	}

	// Return the result
	res.Hash = passwd

	return res, nil
}

func (m *myService) Validate(c context.Context, s *pb.PasswordReq) (*pb.PasswordValidationRes, error) {
	res := &pb.PasswordValidationRes{}

	// Check mandatory fields
	if len(strings.TrimSpace(s.Password)) == 0 {
		res.Error = &pb.Error{
			Code:    http.StatusPreconditionFailed,
			Message: "Password value is mandatory !",
		}
		return res, nil
	}

	if len(strings.TrimSpace(s.Hash)) == 0 {
		res.Error = &pb.Error{
			Code:    http.StatusPreconditionFailed,
			Message: "Hash value is mandatory !",
		}
		return res, nil
	}

	// Hash given password
	valid, err := butcher.Verify([]byte(s.Hash), []byte(s.Password))
	if err != nil {
		res.Error = &pb.Error{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		return res, nil
	}

	// Return result
	res.Valid = valid

	return res, nil
}

func newServer() *myService {
	butch, _ := butcher.New()
	return &myService{
		butch: butch,
	}
}

// -----------------------------------------------------------------------------

func prepareGRPC() (*grpc.Server, error) {
	// gRPC Server settings
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_opentracing.UnaryServerInterceptor(),
				grpc_prometheus.UnaryServerInterceptor,
				// Should always be the last
				panichandler.UnaryPanicHandler,
			),
		),
	)

	// Password service
	pb.RegisterPasswordServer(grpcServer, newServer())

	// Prometheus
	grpc_prometheus.Register(grpcServer)

	// Health
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("Password", healthpb.HealthCheckResponse_SERVING)

	// Reflection
	reflection.Register(grpcServer)

	return grpcServer, nil
}

func prepareHTTP() (*http.Server, error) {
	// Assign a HTTP router
	router := http.NewServeMux()

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

	// gRPC Gateway settings
	ctx := context.Background()
	dopts := []grpc.DialOption{grpc.WithInsecure()}
	gwmux := runtime.NewServeMux()

	// Register Gateway endpoints
	err := pb.RegisterPasswordHandlerFromEndpoint(ctx, gwmux, "localhost:5555", dopts)
	if err != nil {
		return nil, err
	}

	// Assign to router
	router.Handle("/", gwmux)

	// Return HTTP Server instance
	return &http.Server{
		Handler: router,
	}, nil
}

func serve() {
	// Initialize listener
	conn, err := net.Listen("tcp", ":5555")
	if err != nil {
		panic(err)
	}

	// Create the connection muxer
	mux := cmux.New(conn)

	// Connection dispatcher rules
	grpcL := mux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpL := mux.Match(cmux.HTTP1Fast())

	// Create protocol servers

	// gRPC
	grpcServer, err := prepareGRPC()
	if err != nil {
		log.Fatalf("Unable to create gRPC server : %s", err.Error())
	}

	// HTTP
	httpServer, err := prepareHTTP()
	if err != nil {
		log.Fatalf("Unable to create HTTP server : %s", err.Error())
	}

	// Start all muxed listeners
	go grpcServer.Serve(grpcL)
	go httpServer.Serve(httpL)

	if err := mux.Serve(); err != nil {
		log.Fatalf("Unable to serve services, %s", err.Error())
	}
}

package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"go.zenithar.org/common/web/utils"
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

type myService struct{}

func (m *myService) Encode(c context.Context, s *pb.PasswordReq) (*pb.EncodedPasswordRes, error) {
	res := &pb.EncodedPasswordRes{}
	return res, nil
}

func (m *myService) Validate(c context.Context, s *pb.PasswordReq) (*pb.PasswordValidationRes, error) {
	res := &pb.PasswordValidationRes{}
	return res, nil
}

func newServer() *myService {
	return new(myService)
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func serve() {
	// gRPC Server settings
	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPasswordServer(grpcServer, newServer())

	// gRPC Gateway settings
	ctx := context.Background()
	dopts := []grpc.DialOption{grpc.WithInsecure()}
	gwmux := runtime.NewServeMux()
	err := pb.RegisterPasswordHandlerFromEndpoint(ctx, gwmux, "localhost:5555", dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return
	}

	// Assign a HTTP router
	mux := http.NewServeMux()

	// Health monitoring endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		utils.JSONResponse(w, http.StatusOK, map[string]interface{}{
			"status":    "OK",
			"timestamp": time.Now().UTC().Unix(),
		})
	})

	// Service discovery
	mux.HandleFunc("/.well-known/finger", func(w http.ResponseWriter, r *http.Request) {
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
			"endpoints":           []string{"grpc", "http"},
		})
	})

	mux.Handle("/", gwmux)

	// Initialize listener
	conn, err := net.Listen("tcp", ":5555")
	if err != nil {
		panic(err)
	}

	// Initialize a Web Server
	srv := &http.Server{
		Handler: grpcHandlerFunc(grpcServer, mux),
	}

	fmt.Printf("grpc on port: 5555")
	err = srv.Serve(conn)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return
}

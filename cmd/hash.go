package cmd

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	pb "go.zenithar.org/password/protocol/password"
)

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "hash the given password",
	Run: func(cmd *cobra.Command, args []string) {
		// gRPC dialup options
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithTimeout(3*time.Second))
		opts = append(opts, grpc.WithBlock())
		opts = append(opts, grpc.WithInsecure())

		// Initialize connection
		conn, err := grpc.Dial("localhost:5555", opts...)
		if err != nil {
			grpclog.Fatalf("fail to dial: %v", err)
		}
		defer conn.Close()

		// Client stub
		client := pb.NewPasswordClient(conn)

		// Do the call
		res, _ := client.Encode(context.Background(), &pb.PasswordReq{
			Password: strings.Join(os.Args[2:], " "),
		})
		println(res.Hash)
	},
}

func init() {
	RootCmd.AddCommand(hashCmd)
}

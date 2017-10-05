package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	pb "go.zenithar.org/password/protocol/password"
)

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "hash the given password",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn := grpcClientConnection(ctx, "localhost:5555")
		defer conn.Close()

		// Client stub
		client := pb.NewPasswordClient(conn)

		// Do the call
		res, err := client.Encode(ctx, &pb.PasswordReq{
			Password: strings.Join(os.Args[2:], " "),
		})
		if err != nil {
			logrus.WithError(err).Fatal("Unable to do the gRPC call")
		}

		println(res.Hash)
	},
}

func init() {
	RootCmd.AddCommand(hashCmd)
}

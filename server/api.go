package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	pb "go.zenithar.org/password/protocol/password"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"go.zenithar.org/butcher"
)

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

func (m *myService) Ping(c context.Context, s *empty.Empty) (*pb.PongRes, error) {
	ts, _ := ptypes.TimestampProto(time.Now().UTC())
	return &pb.PongRes{
		Timestamp: ts,
	}, nil
}

func newServer() *myService {
	butch, _ := butcher.New()
	return &myService{
		butch: butch,
	}
}

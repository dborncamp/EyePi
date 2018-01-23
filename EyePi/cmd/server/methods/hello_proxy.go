package methods

import (
	"golang.org/x/net/context"

	"github.com/pkg/errors"

	pb "github.com/dborncamp/EyePi/EyePi/protobuf"
)

// HelloProxy says "hello" in a form that is handled by the gateway proxy
func (s *serverData) HelloProxy(_ context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	if req.HelloText == "ping" {
		return &pb.HelloResponse{Text: "pong"}, nil
	}
	return nil, errors.New("invalid request")
}

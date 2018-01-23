package methods

import (
	pb "github.com/dborncamp/EyePi/EyePi/protobuf"
)

type serverData struct {
}

// NewEyePiServer returns an object that implements the pb.EyePiServer interface
func NewEyePiServer() (pb.EyePiServer, error) {
	return &serverData{}, nil
}

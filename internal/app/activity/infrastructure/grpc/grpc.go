package grpc

import (
	"google.golang.org/grpc"

	"github.com/artistomin/proxy/internal/app/activity/usecase/request"
	"github.com/artistomin/proxy/internal/app/activity/usecase/response"
	pb "github.com/artistomin/proxy/internal/pkg/proto/activity"
)

type server struct {
	reqSvc request.Service
	resSvc response.Service
}

// NewServer creates grpc server with handlers
func NewServer(reqSvc request.Service, resSvc response.Service) *grpc.Server {
	server := &server{reqSvc, resSvc}
	s := grpc.NewServer()
	pb.RegisterActivityServer(s, server)

	return s
}

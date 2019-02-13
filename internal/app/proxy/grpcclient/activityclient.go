package grpcclient

import (
	"google.golang.org/grpc"

	pb "github.com/artistomin/proxy/internal/pkg/proto/activity"
)

// NewActivityClient creates new grpc client for activity service
func NewActivityClient(activityAddress string) (*grpc.ClientConn, pb.ActivityClient, error) {
	conn, err := grpc.Dial(activityAddress, grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}

	c := pb.NewActivityClient(conn)

	return conn, c, nil
}

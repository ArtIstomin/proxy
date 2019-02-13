package grpc

import (
	"context"
	"time"

	"github.com/artistomin/proxy/internal/app/activity"
	pb "github.com/artistomin/proxy/internal/pkg/proto/activity"
	"github.com/golang/protobuf/ptypes/empty"
)

func (s *server) StoreRequest(ctx context.Context, in *pb.ReqRequest) (*empty.Empty, error) {
	payload := &activity.Request{
		Request: in.Request,
		URL:     in.Url,
	}

	_, err := s.reqSvc.CreateRequest(payload)

	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *server) UpdateRequest(ctx context.Context, in *pb.ReqRequest) (*empty.Empty, error) {
	reqID := activity.ReqID(in.ReqId)
	payload := &activity.Request{
		ID:        reqID,
		Completed: in.Completed,
		UpdatedAt: time.Now().UTC(),
	}

	err := s.reqSvc.UpdateRequest(reqID, payload)

	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

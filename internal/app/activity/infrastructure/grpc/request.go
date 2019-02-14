package grpc

import (
	"context"
	"time"

	"github.com/artistomin/proxy/internal/app/activity"
	pb "github.com/artistomin/proxy/internal/pkg/proto/activity"
	"github.com/golang/protobuf/ptypes/empty"
)

func (s *server) StoreRequest(ctx context.Context, in *pb.ReqRequest) (*pb.ReqReply, error) {
	payload := &activity.Request{
		Header: in.Header,
		URL:    in.Url,
	}

	id, err := s.reqSvc.CreateRequest(payload)

	if err != nil {
		return nil, err
	}

	return &pb.ReqReply{ReqId: int32(id)}, nil
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

func (s *server) GetRequests(ctx context.Context, in *empty.Empty) (*pb.GetRequestsReply, error) {
	requests, err := s.reqSvc.GetRequests()

	if err != nil {
		return nil, err
	}

	var pbReqs []*pb.ReqRequest

	for _, v := range requests {
		pbReq := &pb.ReqRequest{
			Url:    v.URL,
			Header: v.Header,
		}
		pbReqs = append(pbReqs, pbReq)
	}

	pbRes := &pb.GetRequestsReply{Requests: pbReqs}

	return pbRes, nil
}

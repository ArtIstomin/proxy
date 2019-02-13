package grpc

import (
	"context"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/artistomin/proxy/internal/app/activity"
	pb "github.com/artistomin/proxy/internal/pkg/proto/activity"
)

func (s *server) StoreResponse(ctx context.Context, in *pb.ResRequest) (*empty.Empty, error) {
	exp, _ := ptypes.Timestamp(in.Expires)

	payload := &activity.Response{
		Host:     in.Host,
		URL:      in.Url,
		Body:     in.Body,
		Response: in.Response,
		Expires:  exp,
	}

	err := s.resSvc.CreateOrUpdateResponse(payload)

	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *server) GetResponse(ctx context.Context, in *pb.GetResRequest) (*pb.ResRequest, error) {
	res, err := s.resSvc.GetResponse(in.Host, in.Url)

	if err != nil {
		return nil, err
	}

	exp, _ := ptypes.TimestampProto(res.Expires)
	pbRes := &pb.ResRequest{
		Host:     res.Host,
		Url:      res.URL,
		Response: res.Response,
		Body:     res.Body,
		Expires:  exp,
	}

	return pbRes, nil
}

func (s *server) GetHostSize(ctx context.Context, in *pb.HostSizeRequest) (*pb.HostSizeReply, error) {
	size, err := s.resSvc.GetHostSize(in.Host)

	if err != nil {
		return nil, err
	}

	return &pb.HostSizeReply{Size: size}, nil
}

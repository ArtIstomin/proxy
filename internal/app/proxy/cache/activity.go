package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	pb "github.com/artistomin/proxy/internal/pkg/proto/activity"
	"github.com/golang/protobuf/ptypes"
)

type ActivityCache struct {
	client pb.ActivityClient
}

func (ac *ActivityCache) Get(host, url string) Value {
	payload := &pb.GetResRequest{
		Host: host,
		Url:  url,
	}

	res, err := ac.client.GetResponse(context.Background(), payload)

	if err != nil {
		log.Printf("get cache error: %s", err)
	}

	var response Response
	_ = json.Unmarshal(res.Response, &response)

	return Value{
		Body:     res.Body,
		Response: response,
	}
}

func (ac *ActivityCache) Put(host, url string, response Response, body []byte,
	expires time.Time) {
	respBytes, _ := json.Marshal(response)
	exp, _ := ptypes.TimestampProto(expires)
	payload := &pb.ResRequest{
		Host:     host,
		Url:      url,
		Body:     body,
		Response: respBytes,
		Expires:  exp,
	}

	_, err := ac.client.StoreResponse(context.Background(), payload)

	if err != nil {
		log.Printf("cache error: %s", err)
	}
}

func (ac *ActivityCache) Has(host, key string) bool {
	payload := &pb.GetResRequest{
		Host: host,
		Url:  key,
	}

	res, err := ac.client.GetResponse(context.Background(), payload)

	if err != nil {
		return false
	}

	expires, _ := ptypes.Timestamp(res.Expires)
	now := time.Now().UTC()

	if now.After(expires) {
		return false
	}

	return true
}

func (ac *ActivityCache) Size(host string) int {
	payload := &pb.HostSizeRequest{
		Host: host,
	}

	sizeReply, err := ac.client.GetHostSize(context.Background(), payload)

	if err != nil {
		log.Printf("size cache error: %s", err)
	}

	return int(sizeReply.Size)
}

func NewActivity(grpcClient pb.ActivityClient) *ActivityCache {
	return &ActivityCache{grpcClient}
}

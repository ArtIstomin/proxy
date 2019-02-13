package response

import (
	"time"

	"github.com/artistomin/proxy/internal/app/activity"
	"github.com/jinzhu/gorm"
)

// Service is the interface that provides response methods
type Service interface {
	CreateResponse(res *activity.Response) (activity.ResID, error)
	GetResponse(host, url string) (*activity.Response, error)
	CreateOrUpdateResponse(res *activity.Response) error
	GetHostSize(host string) (int64, error)
}

type service struct {
	resps activity.ResponseRepo
}

func (s *service) CreateResponse(res *activity.Response) (activity.ResID, error) {
	id, err := s.resps.Create(res)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) GetResponse(host, url string) (*activity.Response, error) {
	res, err := s.resps.Get(host, url)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *service) CreateOrUpdateResponse(res *activity.Response) error {
	resp, err := s.resps.Get(res.Host, res.URL)

	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	if resp == nil {
		_, err := s.resps.Create(res)
		if err != nil {
			return err
		}

		return nil
	}

	resp.Response = res.Response
	resp.Body = res.Body
	resp.Expires = res.Expires
	resp.UpdatedAt = time.Now().UTC()

	return s.resps.Update(resp)
}

func (s *service) GetHostSize(host string) (int64, error) {
	return s.resps.GetHostSize(host)
}

// NewService creates a response service with dependencies
func NewService(resps activity.ResponseRepo) Service {
	return &service{resps}
}

package request

import "github.com/artistomin/proxy/internal/app/activity"

// Service is the interface that provides request methods
type Service interface {
	CreateRequest(req *activity.Request) (activity.ReqID, error)
	UpdateRequest(id activity.ReqID, req *activity.Request) error
	GetRequests() ([]*activity.Request, error)
}

type service struct {
	reqs activity.RequestRepo
}

func (s *service) CreateRequest(req *activity.Request) (activity.ReqID, error) {
	id, err := s.reqs.Create(req)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *service) UpdateRequest(id activity.ReqID, req *activity.Request) error {
	err := s.reqs.Update(id, req)

	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetRequests() ([]*activity.Request, error) {
	return s.reqs.GetRequests()
}

// NewService creates a request service with dependencies
func NewService(reqs activity.RequestRepo) Service {
	return &service{reqs}
}

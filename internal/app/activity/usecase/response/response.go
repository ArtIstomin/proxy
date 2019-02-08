package response

import "github.com/artistomin/proxy/internal/app/activity"

// Service is the interface that provides response methods
type Service interface {
	CreateResponse(res *activity.Response) (activity.ResID, error)
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

// NewService creates a response service with dependencies
func NewService(resps activity.ResponseRepo) Service {
	return &service{resps}
}

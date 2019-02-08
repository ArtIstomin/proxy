package postgresql

import (
	"github.com/jinzhu/gorm"

	"github.com/artistomin/proxy/internal/app/activity"
)

type requestRepo struct {
	db *gorm.DB
}

func (rr *requestRepo) Create(req *activity.Request) (activity.ReqID, error) {
	return 0, nil
}

func (rr *requestRepo) Update(id activity.ReqID, req *activity.Request) error {
	return nil
}

// NewRequestRepo returns a new instance of a Postgresql request repository
func NewRequestRepo(db *gorm.DB) activity.RequestRepo {
	return &requestRepo{db}
}

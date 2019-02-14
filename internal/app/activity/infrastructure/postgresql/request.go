package postgresql

import (
	"github.com/jinzhu/gorm"

	"github.com/artistomin/proxy/internal/app/activity"
)

type requestRepo struct {
	db *gorm.DB
}

func (rr *requestRepo) Create(req *activity.Request) (activity.ReqID, error) {
	if err := rr.db.Create(req).Error; err != nil {
		return 0, err
	}

	return req.ID, nil
}

func (rr *requestRepo) Update(id activity.ReqID, req *activity.Request) error {
	if err := rr.db.Model(req).Updates(req).Error; err != nil {
		return err
	}

	return nil
}

func (rr *requestRepo) GetRequests() ([]*activity.Request, error) {
	var requests []*activity.Request

	if err := rr.db.Where("completed = ?", false).Find(&requests).Error; err != nil {
		return nil, err
	}

	return requests, nil
}

// NewRequestRepo returns a new instance of a Postgresql request repository
func NewRequestRepo(db *gorm.DB) activity.RequestRepo {
	return &requestRepo{db}
}

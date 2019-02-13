package postgresql

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/artistomin/proxy/internal/app/activity"
)

type responseRepo struct {
	db *gorm.DB
}

type hostSize struct {
	Size int64
}

func (rr *responseRepo) Create(res *activity.Response) (activity.ResID, error) {
	if err := rr.db.Create(res).Error; err != nil {
		return 0, err
	}

	return res.ID, nil
}

func (rr *responseRepo) Get(host, url string) (*activity.Response, error) {
	resp := new(activity.Response)

	if err := rr.db.Where(&activity.Response{Host: host, URL: url}).Last(&resp).Error; err != nil {
		return nil, err
	}

	return resp, nil
}

func (rr *responseRepo) Update(res *activity.Response) error {
	if err := rr.db.Model(res).Updates(res).Error; err != nil {
		return err
	}

	return nil
}

func (rr *responseRepo) GetHostSize(host string) (int64, error) {
	hs := new(hostSize)
	now := time.Now().UTC()

	err := rr.db.
		Table("responses").
		Select("sum(pg_column_size(response) + pg_column_size(body)) AS size").
		Where("host = ? AND expires > ?", host, now).
		Scan(&hs).
		Error

	if err != nil {
		return 0, err
	}

	return hs.Size, nil
}

// NewResponseRepo returns a new instance of a Postgresql response repository
func NewResponseRepo(db *gorm.DB) activity.ResponseRepo {
	return &responseRepo{db}
}

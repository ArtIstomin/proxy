package postgresql

import (
	"database/sql"

	"github.com/artistomin/proxy/internal/app/activity"
)

type responseRepo struct {
	db *sql.DB
}

func (rr *responseRepo) Create(res *activity.Response) (activity.ResID, error) {
	return 0, nil
}

// NewResponseRepo returns a new instance of a Postgresql response repository
func NewResponseRepo(db *sql.DB) activity.ResponseRepo {
	return &responseRepo{db}
}

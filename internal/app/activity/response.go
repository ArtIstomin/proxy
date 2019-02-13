package activity

import "time"

// ResID uniquely identifies a particular response.
type ResID int32

// Response is central struct in the domain model
type Response struct {
	ID        ResID  `gorm:"primary_key`
	Host      string `gorm:"index:idx_host_url"`
	URL       string `gorm:"index:idx_host_url"`
	Body      []byte
	Response  []byte
	Expires   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ResponseRepo provides access to response entity
type ResponseRepo interface {
	Create(res *Response) (ResID, error)
	Get(host, url string) (*Response, error)
	Update(res *Response) error
	GetHostSize(host string) (int64, error)
}

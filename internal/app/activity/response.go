package activity

import "time"

// ResID uniquely identifies a particular response.
type ResID int32

// Response is central struct in the domain model
type Response struct {
	ID        ResID `gorm:"primary_key`
	ReqID     ReqID
	Response  []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ResponseRepo provides access to response entity
type ResponseRepo interface {
	Create(res *Response) (ResID, error)
}

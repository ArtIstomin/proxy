package activity

import "time"

// ReqID uniquely identifies a particular request.
type ReqID int32

// Request is central struct in the domain model
type Request struct {
	ID        ReqID `gorm:"primary_key`
	URL       string
	Request   []byte
	Completed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RequestRepo provides access to request entity
type RequestRepo interface {
	Create(req *Request) (ReqID, error)
	Update(id ReqID, req *Request) error
}

package postgresql

import (
	"time"

	"github.com/artistomin/proxy/internal/app/activity"
	"github.com/jinzhu/gorm"

	// Initialize postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const (
	maxOpenConn = 5
	maxIdleConn = 5
	maxLifetime = 10 * time.Minute
)

// New creates connection to DB
func New(dsn string) (*gorm.DB, error) {
	gdb, err := gorm.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	db := gdb.DB()
	db.SetMaxOpenConns(maxOpenConn)
	db.SetMaxIdleConns(maxIdleConn)
	db.SetConnMaxLifetime(maxLifetime)

	gdb.AutoMigrate(&activity.Request{}, &activity.Response{})

	return gdb, nil
}

package retry

import (
	"net/http"

	"github.com/artistomin/proxy/internal/app/proxy/cache"
	"github.com/artistomin/proxy/internal/app/proxy/config"
	pb "github.com/artistomin/proxy/internal/pkg/proto/activity"
)

type Retry struct {
	Cache  cache.Cacher
	Tr     *http.Transport
	Config *config.Config
	client pb.ActivityClient
}

type request struct {
}

func (r *Retry) ProcessUncompletedReqs() {
	// http.NewRequest("GET", url, nil)
}

// New returns new retrier
func New(cache cache.Cacher, tr *http.Transport, cfg *config.Config, client pb.ActivityClient) *Retry {
	return &Retry{cache, tr, cfg, client}
}

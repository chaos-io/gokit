package lb

import (
	"crypto/rand"
	"math/big"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
)

// NewRandom returns a load balancer that selects services randomly.
func NewRandom(s sd.Endpointer) Balancer {
	return &random{s: s}
}

type random struct {
	s sd.Endpointer
}

func (r *random) Endpoint() (endpoint.Endpoint, error) {
	endpoints, err := r.s.Endpoints()
	if err != nil {
		return nil, err
	}
	if len(endpoints) <= 0 {
		return nil, ErrNoEndpoints
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(endpoints))))
	if err != nil {
		return nil, err
	}

	return endpoints[n.Int64()], nil
}

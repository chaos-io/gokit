package lb

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
)

func TestRandom(t *testing.T) {
	var (
		n          = 7
		endpoints  = make([]endpoint.Endpoint, n)
		counts     = make([]int, n)
		iterations = 1000000
		want       = iterations / n
		tolerance  = want / 100 // 1%
	)

	for i := 0; i < n; i++ {
		i0 := i
		endpoints[i] = func(context.Context, interface{}) (interface{}, error) { counts[i0]++; return struct{}{}, nil }
	}

	endpointer := sd.FixedEndpointer(endpoints)
	balancer := NewRandom(endpointer)

	for i := 0; i < iterations; i++ {
		_endpoint, _ := balancer.Endpoint()
		_, _ = _endpoint(context.Background(), struct{}{})
	}

	for i, have := range counts {
		delta := int(math.Abs(float64(want - have)))
		if delta > tolerance {
			t.Errorf("%d: want %d, have %d, delta %d > %d tolerance", i, want, have, delta, tolerance)
		}
	}
}

func TestRandomNoEndpoints(t *testing.T) {
	endpointer := sd.FixedEndpointer{}
	balancer := NewRandom(endpointer)
	_, err := balancer.Endpoint()
	if !errors.Is(err, ErrNoEndpoints) {
		t.Errorf("want ErrNoEndpoints, got %v", err)
	}
}

func BenchmarkRandom(b *testing.B) {
	endpointer := sd.FixedEndpointer{}
	nr := NewRandom(endpointer)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = nr.Endpoint()
		}
	})
}

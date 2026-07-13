package grpc

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	kitsd "github.com/go-kit/kit/sd"
	grpcresolver "google.golang.org/grpc/resolver"
)

func TestInstancerResolverUpdatesAndStops(t *testing.T) {
	instancer := &fakeInstancer{}
	clientConn := &fakeResolverClientConn{
		updates: make(chan grpcresolver.State, 1),
		errors:  make(chan error, 1),
	}
	builder := newInstancerBuilder(instancer)

	resolver, err := builder.Build(grpcresolver.Target{}, clientConn, grpcresolver.BuildOptions{})
	if err != nil {
		t.Fatalf("build resolver: %v", err)
	}

	instancer.send(kitsd.Event{Instances: []string{
		" grpc://127.0.0.1:9001 ",
		"127.0.0.1:9000",
		"grpc://127.0.0.1:9001",
		"localhost:9002",
	}})

	select {
	case state := <-clientConn.updates:
		got := make([]string, 0, len(state.Addresses))
		for _, address := range state.Addresses {
			got = append(got, address.Addr)
		}
		want := []string{"127.0.0.1:9000", "127.0.0.1:9001", "localhost:9002"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("addresses: got %v, want %v", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for resolver update")
	}

	discoveryErr := errors.New("discovery unavailable")
	instancer.send(kitsd.Event{Err: discoveryErr})
	select {
	case got := <-clientConn.errors:
		if !errors.Is(got, discoveryErr) {
			t.Fatalf("reported error: got %v, want %v", got, discoveryErr)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for resolver error")
	}

	resolver.Close()
	if !instancer.deregistered() {
		t.Fatal("resolver did not deregister from instancer")
	}
}

type fakeInstancer struct {
	mu                sync.Mutex
	ch                chan<- kitsd.Event
	deregisteredValue bool
}

func (f *fakeInstancer) Register(ch chan<- kitsd.Event) {
	f.mu.Lock()
	f.ch = ch
	f.mu.Unlock()
}

func (f *fakeInstancer) Deregister(ch chan<- kitsd.Event) {
	f.mu.Lock()
	if f.ch == ch {
		f.deregisteredValue = true
		f.ch = nil
	}
	f.mu.Unlock()
}

func (*fakeInstancer) Stop() {}

func (f *fakeInstancer) send(event kitsd.Event) {
	f.mu.Lock()
	ch := f.ch
	f.mu.Unlock()
	ch <- event
}

func (f *fakeInstancer) deregistered() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.deregisteredValue
}

type fakeResolverClientConn struct {
	grpcresolver.ClientConn
	updates chan grpcresolver.State
	errors  chan error
}

func (f *fakeResolverClientConn) UpdateState(state grpcresolver.State) error {
	f.updates <- state
	return nil
}

func (f *fakeResolverClientConn) ReportError(err error) {
	f.errors <- err
}

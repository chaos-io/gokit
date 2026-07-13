package grpc

import (
	"errors"
	"net/url"
	"sort"
	"strings"
	"sync"

	kitsd "github.com/go-kit/kit/sd"
	grpcresolver "google.golang.org/grpc/resolver"
)

const instancerResolverScheme = "gokit"

type instancerBuilder struct {
	instancer kitsd.Instancer
}

func newInstancerBuilder(instancer kitsd.Instancer) *instancerBuilder {
	return &instancerBuilder{instancer: instancer}
}

func (b *instancerBuilder) Scheme() string {
	return instancerResolverScheme
}

func (b *instancerBuilder) Build(_ grpcresolver.Target, clientConn grpcresolver.ClientConn, _ grpcresolver.BuildOptions) (grpcresolver.Resolver, error) {
	if b.instancer == nil {
		return nil, errors.New("grpc: nil service instancer")
	}
	resolver := &instancerResolver{
		instancer:  b.instancer,
		clientConn: clientConn,
		events:     make(chan kitsd.Event, 1),
		done:       make(chan struct{}),
	}
	b.instancer.Register(resolver.events)
	go resolver.watch()
	return resolver, nil
}

type instancerResolver struct {
	instancer  kitsd.Instancer
	clientConn grpcresolver.ClientConn
	events     chan kitsd.Event
	done       chan struct{}
	closeOnce  sync.Once
}

func (r *instancerResolver) watch() {
	for {
		select {
		case event := <-r.events:
			if event.Err != nil {
				r.clientConn.ReportError(event.Err)
				continue
			}
			if err := r.clientConn.UpdateState(grpcresolver.State{Addresses: resolverAddresses(event.Instances)}); err != nil {
				r.clientConn.ReportError(err)
			}
		case <-r.done:
			return
		}
	}
}

func (r *instancerResolver) ResolveNow(grpcresolver.ResolveNowOptions) {}

func (r *instancerResolver) Close() {
	r.closeOnce.Do(func() {
		r.instancer.Deregister(r.events)
		close(r.done)
	})
}

func resolverAddresses(instances []string) []grpcresolver.Address {
	unique := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		if address := resolverAddress(instance); address != "" {
			unique[address] = struct{}{}
		}
	}

	addresses := make([]string, 0, len(unique))
	for address := range unique {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)

	resolved := make([]grpcresolver.Address, 0, len(addresses))
	for _, address := range addresses {
		resolved = append(resolved, grpcresolver.Address{Addr: address})
	}
	return resolved
}

func resolverAddress(instance string) string {
	instance = strings.TrimSpace(instance)
	if instance == "" {
		return ""
	}
	parsed, err := url.Parse(instance)
	if err == nil && strings.Contains(instance, "://") {
		if parsed.Host != "" {
			return parsed.Host
		}
		return strings.TrimPrefix(parsed.Path, "//")
	}
	return instance
}

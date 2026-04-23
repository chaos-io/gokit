package ctxcache

import (
	"context"

	"github.com/chaos-io/chaos/ctxcache"
	"github.com/go-kit/kit/endpoint"
)

func CtxCacheMW(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req any) (resp any, err error) {
		return next(ctxcache.Init(ctx), req)
	}
}

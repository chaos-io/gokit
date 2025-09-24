package ctxcache

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	// "github.com/chaos-io/chaos/pkg/ctxcache"
)

func CtxCacheMW(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req any) (resp any, err error) {
		return next(ctx, req)
		// return next(ctxcache.Init(ctx), req)
	}
}

package validator

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"

	"github.com/go-playground/validator/v10"
)

// 创建全局 validator 实例（线程安全，可复用）
var validate = validator.New()

// ValidatorMW 会自动根据 struct tag 验证请求结构体
func ValidatorMW() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if err := validate.Struct(request); err != nil {
				return nil, fmt.Errorf("failed to validate request: %w", err)
			}
			return next(ctx, request)
		}
	}
}

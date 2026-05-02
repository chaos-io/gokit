package session

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
)

func ValidateMiddleware(validator Validator) endpoint.Middleware {
	if validator == nil {
		return errorMiddleware(ErrValidatorRequired)
	}

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req any) (any, error) {
			token, ok := TokenFromContext(ctx)
			if !ok {
				return nil, ErrTokenRequired
			}

			session, err := validator.Validate(ctx, token)
			if err != nil {
				return nil, fmt.Errorf("validate session: %w", err)
			}

			return next(WithSession(ctx, session), req)
		}
	}
}

func AuthenticateMiddleware(validator Validator, resolver UserResolver) endpoint.Middleware {
	if resolver == nil {
		return errorMiddleware(ErrUserResolverRequired)
	}

	validate := ValidateMiddleware(validator)

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return validate(func(ctx context.Context, req any) (any, error) {
			session, ok := SessionFromContext(ctx)
			if !ok {
				return nil, ErrSessionNotFound
			}

			resolved, err := resolver.ResolveUser(ctx, session, req)
			if err != nil {
				return nil, fmt.Errorf("resolve user: %w", err)
			}
			if resolved == nil || resolved.ID == "" || resolved.Value == nil {
				return nil, ErrResolvedUserInvalid
			}
			if resolved.ID != session.Subject.UserID {
				return nil, ErrResolvedUserMismatch
			}

			return next(WithUser(ctx, resolved.Value), req)
		})
	}
}

func errorMiddleware(err error) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req any) (any, error) {
			return nil, err
		}
	}
}

package locale

import (
	"context"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/stretchr/testify/require"
)

func TestWithLocale(t *testing.T) {
	ctx := WithLocale(context.Background(), " zh_CN,zh;q=0.9 ")

	locale, ok := LocaleFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "zh-cn", locale)
}

func TestNewLocaleMW(t *testing.T) {
	ep := NewLocaleMW(ResolverFunc(func(ctx context.Context, req any) (string, bool, error) {
		return "en_US", true, nil
	}))(func(ctx context.Context, request any) (any, error) {
		locale, ok := LocaleFromContext(ctx)
		require.True(t, ok)
		require.Equal(t, "en-us", locale)
		return "ok", nil
	})

	response, err := ep(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, "ok", response)
}

func TestNewLocaleMWUsesExistingContextLocale(t *testing.T) {
	ep := NewLocaleMW(nil)(endpoint.Endpoint(func(ctx context.Context, request any) (any, error) {
		locale, ok := LocaleFromContext(ctx)
		require.True(t, ok)
		require.Equal(t, "fr-fr", locale)
		return nil, nil
	}))

	_, err := ep(WithLocale(context.Background(), "fr_FR"), nil)
	require.NoError(t, err)
}

func TestResolveFromBase(t *testing.T) {
	type base struct {
		Locale string
	}

	type request struct {
		Base *base
	}

	localeValue, ok := ResolveFromBase(&request{
		Base: &base{Locale: " zh_CN "},
	})
	require.True(t, ok)
	require.Equal(t, "zh_CN", localeValue)

	_, ok = ResolveFromBase(&request{})
	require.False(t, ok)
}

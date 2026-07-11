package pagination

const (
	DefaultPageSize = 50
	MaxPageSize     = 1000
)

// NormalizePageSize applies the API defaults and limits for a list request.
func NormalizePageSize(pageSize int) (int, error) {
	if pageSize < 0 {
		return 0, ErrInvalidPageSize
	}

	if pageSize == 0 {
		return DefaultPageSize, nil
	}

	if pageSize > MaxPageSize {
		return MaxPageSize, nil
	}
	return pageSize, nil
}

// Page is a standard list result. It satisfies Paginator so transports can
// consistently expose pagination metadata without knowing the concrete item type.
type Page[T any] struct {
	Items         []T
	TotalCount    int32
	NextPageToken string
}

func (p Page[T]) GetTotalCount() int32     { return p.TotalCount }
func (p Page[T]) GetNextPageToken() string { return p.NextPageToken }

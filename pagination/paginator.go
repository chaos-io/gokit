package pagination

type Paginator interface {
	GetTotalCount() int32
	GetNextPageToken() string
}

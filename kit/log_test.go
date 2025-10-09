package kit

import (
	"testing"

	"github.com/chaos-io/chaos/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func Test_kitLogger_Log(t *testing.T) {
	l := kitLogger{
		Logger: logs.DefaultLogger(),
	}

	kvs := []any{"1", 1}
	err := l.Log(kvs...)
	assert.NoError(t, err)
}

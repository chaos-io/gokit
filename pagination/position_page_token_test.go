package pagination

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPositionPageToken_Create(t *testing.T) {
	token := (&PositionPageToken{}).Create(100).Format()
	_ = token

	assert.NotEmpty(t, token)

}

func TestPositionPageToken_Parse(t *testing.T) {
	token := (&PositionPageToken{}).Create(100).Format()
	token2 := &PositionPageToken{}
	err := token2.Parse(token)
	assert.NoError(t, err)
	assert.Equal(t, int32(100), token2.Position)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int32
		isValid  bool
	}{
		{"int - normal", int(100), 100, true},
		{"int - too large", int(int64(math.MaxInt32) + 1), 0, false},

		{"int8", int8(127), 127, true},
		{"int16", int16(30000), 30000, true},
		{"int32", int32(123456), 123456, true},
		{"int64 - within range", int64(200000), 200000, true},
		{"int64 - too large", int64(math.MaxInt64), 0, false},

		{"uint", uint(200000), 200000, true},
		{"uint - too large", uint(math.MaxUint32 + 1), 0, false},
		{"uint8", uint8(255), 255, true},
		{"uint16", uint16(65535), 65535, true},
		{"uint32 - ok", uint32(math.MaxInt32), math.MaxInt32, true},
		{"uint32 - too big", uint32(math.MaxInt32 + 1), 0, false},
		{"uint64 - ok", uint64(100000), 100000, true},
		{"uint64 - too big", uint64(math.MaxUint64), 0, false},

		{"string - invalid", "not a number", 0, false},
		{"nil - invalid", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PositionPageToken{}
			result := p.Create(tt.input)

			token, ok := result.(*PositionPageToken)
			if !tt.isValid {
				if ok && token.Position != 0 {
					t.Errorf("expected invalid token, got: %+v", token)
				}
			} else {
				if !ok {
					t.Errorf("expected valid token, got nil or wrong type")
				}
				if token.Position != tt.expected {
					t.Errorf("expected position %d, got %d", tt.expected, token.Position)
				}
				if time.Since(token.Time) > time.Second {
					t.Errorf("unexpected timestamp: %v", token.Time)
				}
			}
		})
	}
}

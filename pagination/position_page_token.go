package pagination

import (
	"errors"
	"math"
	"time"

	"github.com/mr-tron/base58"
)

// PositionPageToken is a simple cursor-based pagination.
type PositionPageToken struct {
	time.Time
	Position int32
}

func (p *PositionPageToken) Format() string {
	seconds := p.Unix()
	bytes := []byte{
		byte(p.Position >> 16),
		byte(p.Position >> 8),
		byte(p.Position),
		byte(seconds >> 32),
		byte(seconds >> 24),
		byte(seconds >> 16),
		byte(seconds >> 8),
		byte(seconds),
	}
	return "p" + base58.Encode(bytes)
}

func (p *PositionPageToken) Parse(token string) error {
	if len(token) > 0 && token[0] == 'p' {
		token = token[1:]

		bytes, err := base58.Decode(token)
		if err != nil {
			return err
		}

		if len(bytes) < 8 {
			return errors.New("invalid token: insufficient length data")
		}

		var seconds int64
		seconds |= int64(bytes[3]) << 32
		seconds |= int64(bytes[4]) << 24
		seconds |= int64(bytes[5]) << 16
		seconds |= int64(bytes[6]) << 8
		seconds |= int64(bytes[7]) << 0

		var position int64
		position |= int64(bytes[0]) << 16
		position |= int64(bytes[1]) << 8
		position |= int64(bytes[2]) << 0

		p.Time = time.Unix(seconds, 0)
		p.Position = int32(position) // #nosec G115
	}

	return nil
}

func (p *PositionPageToken) Value() interface{} {
	return p.Position
}

func (*PositionPageToken) Create(value interface{}) PageToken {
	var pos int32
	switch v := value.(type) {
	case int:
		if v > math.MaxInt32 || v < math.MinInt32 {
			return &PositionPageToken{}
		}
		pos = int32(v)
	case int8:
		pos = int32(v)
	case int16:
		pos = int32(v)
	case int32:
		pos = v
	case int64:
		if v > math.MaxInt32 || v < math.MinInt32 {
			return &PositionPageToken{}
		}
		pos = int32(v)
	case uint:
		if v > math.MaxInt32 {
			return &PositionPageToken{}
		}
		pos = int32(v)
	case uint8:
		pos = int32(v)
	case uint16:
		pos = int32(v)
	case uint32:
		if v > math.MaxInt32 {
			return &PositionPageToken{}
		}
		pos = int32(v)
	case uint64:
		if v > math.MaxInt32 {
			return &PositionPageToken{}
		}
		pos = int32(v)
	default:
		return &PositionPageToken{}
	}

	return &PositionPageToken{
		Time:     time.Now(),
		Position: pos,
	}
}

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseUrl(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{raw: "http://example.com", want: "http://example.com", wantErr: false},
		{raw: "https://example.com", want: "https://example.com", wantErr: false},
		{raw: "http://example.com/resources/123", want: "http://example.com/resources/123", wantErr: false},
		{raw: "http://example.com/resources?id=123", want: "http://example.com/resources?id=123", wantErr: false},
		{raw: "example.com/resources/123", want: "https://example.com/resources/123", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_url, err := ParseUrl(tt.raw)
			if err != nil {
				t.Logf("ParseUrl(%q): %v\n", tt.raw, err)
			}
			got := _url.String()
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

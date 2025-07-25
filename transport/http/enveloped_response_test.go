package http

import (
	"testing"

	"github.com/chaos-io/core/go/chaos/core"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

func TestEnvelopedResponse_ToErrorWrapped(t *testing.T) {
	resp := &EnvelopedResponse{
		Error:         core.NewErrorFrom(400, "Invalid Arguments"),
		Data:          nil,
		TotalCount:    0,
		NextPageToken: "",
	}

	wer := resp.ToErrorWrapped()
	json, err := jsoniter.Marshal(wer)
	assert.NoError(t, err)

	wrapped := &ErrorWrappedEnvelopedResponse{}
	err = jsoniter.Unmarshal(json, wrapped)
	assert.NoError(t, err)
	assert.NotNil(t, wrapped.Error)
}

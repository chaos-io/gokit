package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHMACCodecRoundTripAndKeyRotation(t *testing.T) {
	oldKeyring, err := NewStaticKeyring(testFallbackKey())
	require.NoError(t, err)

	oldCodec, err := NewHMACCodec(oldKeyring)
	require.NoError(t, err)

	claims := TokenClaims{
		SessionID: "session-1",
		UserID:    "user-1",
		AppID:     7,
		IssuedAt:  100,
		ExpiresAt: 200,
	}

	token, err := oldCodec.Encode(context.Background(), claims)
	require.NoError(t, err)

	newKeyring, err := NewStaticKeyring(testActiveKey(), testFallbackKey())
	require.NoError(t, err)

	newCodec, err := NewHMACCodec(newKeyring)
	require.NoError(t, err)

	decoded, err := newCodec.Decode(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, &claims, decoded)
}

func TestHMACCodecValidatesInput(t *testing.T) {
	_, err := NewStaticKeyring(Key{})
	require.ErrorIs(t, err, ErrSigningKeyIDRequired)

	keyring, err := NewStaticKeyring(testActiveKey())
	require.NoError(t, err)

	codec, err := NewHMACCodec(keyring)
	require.NoError(t, err)

	_, err = codec.Decode(context.Background(), "")
	require.ErrorIs(t, err, ErrTokenRequired)

	_, err = codec.Decode(context.Background(), "bad-token")
	require.ErrorIs(t, err, ErrTokenMalformed)

	_, err = codec.Encode(context.Background(), TokenClaims{})
	require.ErrorIs(t, err, ErrTokenClaimsInvalid)
}

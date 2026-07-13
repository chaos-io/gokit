package grpc

import (
	"testing"
)

func TestNewClientRequiresTransportCredentials(t *testing.T) {
	conn, err := NewClient("user.v1.UserService")
	if err == nil {
		_ = conn.Close()
		t.Fatal("expected missing transport credentials error")
	}
}

func TestNewClientAcceptsExplicitInsecureTransport(t *testing.T) {
	conn, err := NewClient("user.v1.UserService", WithInsecure())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close client: %v", err)
	}
}

package etcdv3

import "testing"

func TestNewServicePreservesInstanceValue(t *testing.T) {
	service, err := newService("http://127.0.0.1:8080", "/users/")
	if err != nil {
		t.Fatal(err)
	}
	if service.Key != "/users/127.0.0.1:8080" {
		t.Fatalf("unexpected key %q", service.Key)
	}
	if service.Value != "http://127.0.0.1:8080" {
		t.Fatalf("unexpected value %q", service.Value)
	}
}

func TestNewServiceAcceptsBareHostPort(t *testing.T) {
	service, err := newService("127.0.0.1:8080", "users")
	if err != nil {
		t.Fatal(err)
	}
	if service.Key != "/users/127.0.0.1:8080" {
		t.Fatalf("unexpected key %q", service.Key)
	}
	if service.Value != "127.0.0.1:8080" {
		t.Fatalf("unexpected value %q", service.Value)
	}
}

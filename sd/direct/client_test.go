package direct

import (
	"reflect"
	"testing"

	kitsd "github.com/go-kit/kit/sd"
)

func TestInstancerReturnsCleanCopy(t *testing.T) {
	cfg := map[string]*Config{
		"users": {Urls: []string{" http://127.0.0.1:8080 ", "", "grpc://127.0.0.1:9000"}},
	}
	client := New(cfg)
	cfg["users"].Urls[0] = "changed"

	instancer := client.Instancer("users")
	fixed, ok := instancer.(kitsd.FixedInstancer)
	if !ok {
		t.Fatalf("expected FixedInstancer, got %T", instancer)
	}

	want := kitsd.FixedInstancer{"http://127.0.0.1:8080", "grpc://127.0.0.1:9000"}
	if !reflect.DeepEqual(want, fixed) {
		t.Fatalf("want %v, got %v", want, fixed)
	}
}

func TestRegisterAndDeregister(t *testing.T) {
	client := New(map[string]*Config{
		"users": {Urls: []string{"http://127.0.0.1:8080"}},
	})

	if err := client.Register(" http://127.0.0.1:8081 ", "users", nil); err != nil {
		t.Fatal(err)
	}
	want := kitsd.FixedInstancer{"http://127.0.0.1:8080", "http://127.0.0.1:8081"}
	if have := fixedInstancer(t, client, "users"); !reflect.DeepEqual(want, have) {
		t.Fatalf("want %v, got %v", want, have)
	}

	if err := client.Deregister(); err != nil {
		t.Fatal(err)
	}
	want = kitsd.FixedInstancer{"http://127.0.0.1:8080"}
	if have := fixedInstancer(t, client, "users"); !reflect.DeepEqual(want, have) {
		t.Fatalf("want %v, got %v", want, have)
	}
}

func TestDeregisterKeepsStaticDuplicate(t *testing.T) {
	client := New(map[string]*Config{
		"users": {Urls: []string{"http://127.0.0.1:8080"}},
	})

	if err := client.Register("http://127.0.0.1:8080", "users", nil); err != nil {
		t.Fatal(err)
	}
	if err := client.Deregister(); err != nil {
		t.Fatal(err)
	}

	want := kitsd.FixedInstancer{"http://127.0.0.1:8080"}
	if have := fixedInstancer(t, client, "users"); !reflect.DeepEqual(want, have) {
		t.Fatalf("want %v, got %v", want, have)
	}
}

func TestRegisterValidatesInput(t *testing.T) {
	client := New(nil)
	if err := client.Register("", "users", nil); err == nil {
		t.Fatal("expected error")
	}
	if err := client.Register("http://127.0.0.1:8080", "", nil); err == nil {
		t.Fatal("expected error")
	}

	var nilClient *Client
	if err := nilClient.Register("http://127.0.0.1:8080", "users", nil); err == nil {
		t.Fatal("expected error")
	}
	if err := nilClient.Deregister(); err != nil {
		t.Fatal(err)
	}

	if err := client.Register("http://127.0.0.1:8080", "users", nil); err != nil {
		t.Fatal(err)
	}
	if err := client.Register("http://127.0.0.1:8081", "orders", nil); err != nil {
		t.Fatal(err)
	}
	if instancer := client.Instancer("users"); instancer != nil {
		t.Fatalf("expected previous registration removed, got %T", instancer)
	}
}

func TestRegisterWorksOnZeroValueClient(t *testing.T) {
	var client Client
	if err := client.Register("http://127.0.0.1:8080", "users", nil); err != nil {
		t.Fatal(err)
	}

	want := kitsd.FixedInstancer{"http://127.0.0.1:8080"}
	if have := fixedInstancer(t, &client, "users"); !reflect.DeepEqual(want, have) {
		t.Fatalf("want %v, got %v", want, have)
	}
}

func fixedInstancer(t *testing.T, client *Client, service string) kitsd.FixedInstancer {
	t.Helper()
	instancer := client.Instancer(service)
	fixed, ok := instancer.(kitsd.FixedInstancer)
	if !ok {
		t.Fatalf("expected FixedInstancer, got %T", instancer)
	}
	return fixed
}

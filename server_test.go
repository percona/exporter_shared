package exporter_shared

import (
	"testing"
)

func Test_getListenerByAddr(t *testing.T) {

	data := struct {
		in  string
		out string
	}{
		in:  "unix:///tmp/exporter_shared_test_socket",
		out: "/tmp/exporter_shared_test_socket",
	}

	listener, _ := getListenerByAddr(data.in)
	if listener.Addr().String() != data.out {
		t.Errorf("getListener(%s).Addr().String() got %s, want %s", data.in, listener.Addr().String(), data.out)
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("Can't close opened listener, reason: %v", err)
	}
}

func Test_parseAddr(t *testing.T) {

	tests := []struct {
		name string
		in   string
		out  struct{ network, address string }
	}{
		{"unknown", "127.0.0.1:2315", struct{ network, address string }{"tcp", "127.0.0.1:2315"}},
		{"empty", "tcp://:2315", struct{ network, address string }{"tcp", ":2315"}},
		{"tcp", "tcp://127.0.0.1:2315", struct{ network, address string }{"tcp", "127.0.0.1:2315"}},
		{"http", "http://127.0.0.1:8080", struct{ network, address string }{"http", "127.0.0.1:8080"}},
		{"https", "https://127.0.0.1:443", struct{ network, address string }{"https", "127.0.0.1:443"}},
		{"unix", "unix:///tmp/unix_socket", struct{ network, address string }{"unix", "/tmp/unix_socket"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			network, address := parseAddr(tt.in)
			if tt.out.network != network {
				t.Errorf("parseAddr(%s) got %s, want %s", tt.in, network, tt.out.network)
			}

			if tt.out.address != address {
				t.Errorf("parseAddr(%s) got %s, want %s", tt.in, address, tt.out.address)
			}
		})
	}

}

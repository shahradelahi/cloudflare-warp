package sdk

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/model"
)

func testIdentity() *model.Identity {
	return &model.Identity{
		PrivateKey: "private-key",
		Config: model.IdentityConfig{
			Peers: []model.IdentityConfigPeer{{PublicKey: "public-key"}},
		},
	}
}

func TestNewProxyDefaults(t *testing.T) {
	client, err := NewClient(ClientConfig{Identity: testIdentity()})
	if err != nil {
		t.Fatal(err)
	}

	proxy, err := client.NewProxy(context.Background(), ProxyConfig{
		Port:       1080,
		EndpointIP: netip.MustParseAddr("162.159.192.1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if got, want := proxy.Addr().String(), "127.0.0.1:1080"; got != want {
		t.Fatalf("Addr() = %q, want %q", got, want)
	}
	if got, want := proxy.Endpoint().String(), "162.159.192.1:2408"; got != want {
		t.Fatalf("Endpoint() = %q, want %q", got, want)
	}
	if err := proxy.Wait(); !errors.Is(err, ErrNotStarted) {
		t.Fatalf("Wait() error = %v, want %v", err, ErrNotStarted)
	}
	proxy.Stop()
}

func TestNewProxyValidation(t *testing.T) {
	client, err := NewClient(ClientConfig{Identity: testIdentity()})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		config ProxyConfig
	}{
		{name: "missing port", config: ProxyConfig{EndpointIP: netip.MustParseAddr("162.159.192.1")}},
		{name: "missing endpoint", config: ProxyConfig{Port: 1080}},
		{name: "invalid protocol", config: ProxyConfig{Port: 1080, EndpointIP: netip.MustParseAddr("162.159.192.1"), Protocol: Protocol(10)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := client.NewProxy(context.Background(), test.config); err == nil {
				t.Fatal("NewProxy() error = nil, want validation error")
			}
		})
	}
}

func TestNewClientValidatesIdentity(t *testing.T) {
	_, err := NewClient(ClientConfig{Identity: &model.Identity{}})
	if err == nil {
		t.Fatal("NewClient() error = nil, want validation error")
	}
}

func TestStopBeforeStart(t *testing.T) {
	client, err := NewClient(ClientConfig{Identity: testIdentity()})
	if err != nil {
		t.Fatal(err)
	}
	proxy, err := client.NewProxy(context.Background(), ProxyConfig{
		Port:       1080,
		EndpointIP: netip.MustParseAddr("162.159.192.1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	proxy.Stop()
	if err := proxy.Start(); err != nil {
		t.Fatal(err)
	}
	if err := proxy.Wait(); err != nil {
		t.Fatalf("Wait() error = %v, want nil", err)
	}
}

func TestGenerateIdentityRequiresDirectory(t *testing.T) {
	if _, err := GenerateIdentity(""); err == nil {
		t.Fatal("GenerateIdentity() error = nil, want validation error")
	}
}

func TestNormalizeScanOptions(t *testing.T) {
	options := normalizeScanOptions(ScanOptions{})
	if !options.IPv4 || !options.IPv6 {
		t.Fatal("normalizeScanOptions() did not enable both address families")
	}
	if options.Limit != defaultScanLimit {
		t.Fatalf("Limit = %d, want %d", options.Limit, defaultScanLimit)
	}
	if options.MaxRTT != defaultMaxRTT {
		t.Fatalf("MaxRTT = %s, want %s", options.MaxRTT, defaultMaxRTT)
	}
	if options.Timeout != defaultScanTimeout {
		t.Fatalf("Timeout = %s, want %s", options.Timeout, defaultScanTimeout)
	}
}

func TestCheckEndpointValidation(t *testing.T) {
	client, err := NewClient(ClientConfig{Identity: testIdentity()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.CheckEndpoint(context.Background(), netip.AddrPort{}); err == nil {
		t.Fatal("CheckEndpoint() error = nil, want validation error")
	}
}

func TestShutdownBeforeStart(t *testing.T) {
	client, err := NewClient(ClientConfig{Identity: testIdentity()})
	if err != nil {
		t.Fatal(err)
	}
	proxy, err := client.NewProxy(context.Background(), ProxyConfig{
		Port:       1080,
		EndpointIP: netip.MustParseAddr("162.159.192.1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := proxy.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v, want nil", err)
	}
}

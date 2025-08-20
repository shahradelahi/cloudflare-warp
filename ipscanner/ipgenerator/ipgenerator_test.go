package ipgenerator

import (
	"net/netip"
	"testing"
)

func TestIpGenerator_SingleCIDR(t *testing.T) {
	cidr, _ := netip.ParsePrefix("192.168.1.0/29") // 8 IPs
	gen, err := NewIpGenerator([]netip.Prefix{cidr})
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	var count int
	ips := make(map[netip.Addr]bool)
	for {
		ip, ok := gen.Next()
		if !ok {
			break
		}
		if !cidr.Contains(ip) {
			t.Errorf("generated IP %s not in CIDR %s", ip, cidr)
		}
		if ips[ip] {
			t.Errorf("duplicate IP generated: %s", ip)
		}
		ips[ip] = true
		count++
	}

	if count != 8 {
		t.Errorf("expected 8 IPs, got %d", count)
	}
}

func TestIpGenerator_MultipleCIDRs(t *testing.T) {
	cidr1, _ := netip.ParsePrefix("10.0.0.0/30") // 4 IPs
	cidr2, _ := netip.ParsePrefix("10.0.1.0/30") // 4 IPs
	gen, err := NewIpGenerator([]netip.Prefix{cidr1, cidr2})
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	var count int
	ips := make(map[netip.Addr]bool)
	for {
		ip, ok := gen.Next()
		if !ok {
			break
		}
		if !cidr1.Contains(ip) && !cidr2.Contains(ip) {
			t.Errorf("generated IP %s not in any CIDR", ip)
		}
		if ips[ip] {
			t.Errorf("duplicate IP generated: %s", ip)
		}
		ips[ip] = true
		count++
	}

	if count != 8 {
		t.Errorf("expected 8 IPs, got %d", count)
	}
}

func TestIpGenerator_NoCIDRs(t *testing.T) {
	gen, err := NewIpGenerator([]netip.Prefix{})
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	_, ok := gen.Next()
	if ok {
		t.Error("expected no IP from an empty generator")
	}
}

func TestIpGenerator_IPv6(t *testing.T) {
	cidr, _ := netip.ParsePrefix("2001:db8::/126") // 4 IPs
	gen, err := NewIpGenerator([]netip.Prefix{cidr})
	if err != nil {
		t.Fatalf("failed to create generator: %v", err)
	}

	var count int
	ips := make(map[netip.Addr]bool)
	for {
		ip, ok := gen.Next()
		if !ok {
			break
		}
		if !cidr.Contains(ip) {
			t.Errorf("generated IP %s not in CIDR %s", ip, cidr)
		}
		if ips[ip] {
			t.Errorf("duplicate IP generated: %s", ip)
		}
		ips[ip] = true
		count++
	}

	if count != 4 {
		t.Errorf("expected 4 IPs, got %d", count)
	}
}

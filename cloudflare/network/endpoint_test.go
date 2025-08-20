package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomScannerPrefix(t *testing.T) {
	// Test IPv4
	prefixV4 := RandomScannerPrefix(true, false)
	assert.True(t, prefixV4.Addr().Is4(), "Expected an IPv4 prefix")
	assert.False(t, prefixV4.Addr().Is6(), "Expected an IPv4 prefix, but got IPv6")

	// Test IPv6
	prefixV6 := RandomScannerPrefix(false, true)
	assert.True(t, prefixV6.Addr().Is6(), "Expected an IPv6 prefix")
	assert.False(t, prefixV6.Addr().Is4(), "Expected an IPv6 prefix, but got IPv4")

	// Test both
	prefixBoth := RandomScannerPrefix(true, true)
	assert.True(t, prefixBoth.Addr().Is4() || prefixBoth.Addr().Is6(), "Expected either an IPv4 or IPv6 prefix")
}

func TestRandomScannerPort(t *testing.T) {
	port := RandomScannerPort()
	assert.Contains(t, ScannerPorts(), port, "RandomScannerPort should return a port from the ScannerPorts list")
}

func TestRandomScannerEndpoint(t *testing.T) {
	// Test IPv4
	endpointV4, err := RandomScannerEndpoint(true, false)
	assert.NoError(t, err)
	assert.True(t, endpointV4.Addr().Is4(), "Expected an IPv4 endpoint")
	assert.Contains(t, ScannerPorts(), endpointV4.Port(), "Endpoint port should be in the ScannerPorts list")

	// Test IPv6
	endpointV6, err := RandomScannerEndpoint(false, true)
	assert.NoError(t, err)
	assert.True(t, endpointV6.Addr().Is6(), "Expected an IPv6 endpoint")
	assert.Contains(t, ScannerPorts(), endpointV6.Port(), "Endpoint port should be in the ScannerPorts list")
}

package network

import (
	"math/rand"
	"net/netip"
	"time"

	"github.com/shahradelahi/cloudflare-warp/utils"
)

func ScannerPrefixes() []netip.Prefix {
	return []netip.Prefix{
		netip.MustParsePrefix("162.159.192.0/24"),
		netip.MustParsePrefix("162.159.195.0/24"),
		netip.MustParsePrefix("188.114.96.0/24"),
		netip.MustParsePrefix("188.114.97.0/24"),
		netip.MustParsePrefix("188.114.98.0/24"),
		netip.MustParsePrefix("188.114.99.0/24"),
		netip.MustParsePrefix("2606:4700:d0::/48"),
		netip.MustParsePrefix("2606:4700:d1::/48"),
	}
}

func RandomScannerPrefix(v4, v6 bool) netip.Prefix {
	if !v4 && !v6 {
		panic("Must choose a IP version for RandomScannerPrefix")
	}

	cidrs := ScannerPrefixes()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		cidr := cidrs[rng.Intn(len(cidrs))]

		if v4 && cidr.Addr().Is4() {
			return cidr
		}

		if v6 && cidr.Addr().Is6() {
			return cidr
		}
	}
}

func ScannerPorts() []uint16 {
	return []uint16{
		443,
		500,
		854,
		859,
		864,
		878,
		880,
		890,
		891,
		894,
		903,
		908,
		928,
		934,
		939,
		942,
		943,
		945,
		946,
		955,
		968,
		987,
		988,
		1002,
		1010,
		1014,
		1018,
		1070,
		1074,
		1180,
		1387,
		1701,
		1843,
		2371,
		2408,
		2506,
		3138,
		3476,
		3581,
		3854,
		4177,
		4198,
		4233,
		4443,
		4500,
		5279,
		5956,
		7103,
		7152,
		7156,
		7281,
		7559,
		8095,
		8319,
		8443,
		8742,
		8854,
		8886,
	}
}

func RandomScannerPort() uint16 {
	ports := ScannerPorts()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return ports[rng.Intn(len(ports))]
}

func RandomScannerEndpoint(v4, v6 bool) (netip.AddrPort, error) {
	randomIP, err := utils.RandomIPFromPrefix(RandomScannerPrefix(v4, v6))
	if err != nil {
		return netip.AddrPort{}, err
	}

	return netip.AddrPortFrom(randomIP, RandomScannerPort()), nil
}

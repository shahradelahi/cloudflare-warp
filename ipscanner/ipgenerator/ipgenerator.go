package ipgenerator

import (
	"math/big"
	"net/netip"
)

// IpGenerator is a new IP address generator.
type IpGenerator struct {
	ipRanges     []IPRange
	currentRange int
}

// NewIpGenerator creates a new IpGenerator.
func NewIpGenerator(cidrs []netip.Prefix) (*IpGenerator, error) {
	var ipRanges []IPRange
	for _, cidr := range cidrs {
		ipRange, err := NewIPRange(cidr)
		if err != nil {
			// We can choose to skip invalid CIDRs or return an error.
			// For now, let's skip.
			continue
		}
		ipRanges = append(ipRanges, ipRange)
	}
	return &IpGenerator{
		ipRanges:     ipRanges,
		currentRange: 0,
	}, nil
}

// Next returns the next IP address in the sequence.
// It iterates through ranges sequentially. When a range is exhausted, it moves to the next.
// When all ranges are exhausted, it returns nil.
func (g *IpGenerator) Next() (netip.Addr, bool) {
	if len(g.ipRanges) == 0 {
		return netip.Addr{}, false
	}

	for g.currentRange < len(g.ipRanges) {
		ip, ok := g.ipRanges[g.currentRange].Next()
		if ok {
			return ip, true
		}
		// Current range is exhausted, move to the next one.
		g.currentRange++
	}

	// All ranges are exhausted.
	return netip.Addr{}, false
}

// GetAll returns all IP addresses from all ranges.
func (g *IpGenerator) GetAll() []netip.Addr {
	var allIPs []netip.Addr
	for _, r := range g.ipRanges {
		allIPs = append(allIPs, r.GetAll()...)
	}
	return allIPs
}

// IPRange represents a range of IP addresses from a CIDR.
type IPRange struct {
	start   *big.Int
	current *big.Int
	end     *big.Int
	isIPv4  bool
}

// NewIPRange creates a new IPRange from a CIDR prefix.
func NewIPRange(cidr netip.Prefix) (IPRange, error) {
	startIP := cidr.Addr()
	startInt := big.NewInt(0).SetBytes(startIP.AsSlice())

	// Calculate the last IP address in the CIDR range.
	prefixLen := cidr.Bits()
	addrLen := 128
	if startIP.Is4() {
		addrLen = 32
	}

	// Create a mask to get the last IP.
	mask := new(big.Int).Lsh(big.NewInt(1), uint(addrLen-prefixLen))
	mask.Sub(mask, big.NewInt(1))
	mask.Not(mask)

	networkInt := new(big.Int).And(startInt, mask)

	broadcastMask := new(big.Int).Lsh(big.NewInt(1), uint(addrLen-prefixLen))
	broadcastMask.Sub(broadcastMask, big.NewInt(1))

	endInt := new(big.Int).Or(networkInt, broadcastMask)

	return IPRange{
		start:   startInt,
		current: new(big.Int).Set(startInt),
		end:     endInt,
		isIPv4:  startIP.Is4(),
	}, nil
}

// Next returns the next IP address in the range.
func (r *IPRange) Next() (netip.Addr, bool) {
	if r.current.Cmp(r.end) > 0 {
		return netip.Addr{}, false
	}

	ipBytes := r.current.Bytes()
	var ip netip.Addr
	var ok bool

	// Pad with leading zeros if necessary
	addrLen := 16 // IPv6
	if r.isIPv4 {
		addrLen = 4
	}
	if len(ipBytes) < addrLen {
		paddedBytes := make([]byte, addrLen)
		copy(paddedBytes[addrLen-len(ipBytes):], ipBytes)
		ip, ok = netip.AddrFromSlice(paddedBytes)
	} else {
		ip, ok = netip.AddrFromSlice(ipBytes)
	}

	if !ok {
		// This should ideally not happen if logic is correct.
		return netip.Addr{}, false
	}

	r.current.Add(r.current, big.NewInt(1))

	if r.isIPv4 {
		return ip.Unmap(), true
	}
	return ip, true
}

// GetAll returns all IP addresses in the range.
func (r *IPRange) GetAll() []netip.Addr {
	var ips []netip.Addr
	current := new(big.Int).Set(r.start)
	for current.Cmp(r.end) <= 0 {
		ipBytes := current.Bytes()
		var ip netip.Addr
		var ok bool

		addrLen := 16 // IPv6
		if r.isIPv4 {
			addrLen = 4
		}
		if len(ipBytes) < addrLen {
			paddedBytes := make([]byte, addrLen)
			copy(paddedBytes[addrLen-len(ipBytes):], ipBytes)
			ip, ok = netip.AddrFromSlice(paddedBytes)
		} else {
			ip, ok = netip.AddrFromSlice(ipBytes)
		}

		if ok {
			if r.isIPv4 {
				ips = append(ips, ip.Unmap())
			} else {
				ips = append(ips, ip)
			}
		}
		current.Add(current, big.NewInt(1))
	}
	return ips
}

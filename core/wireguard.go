package core

import (
	"net/netip"

	"github.com/shahradelahi/wiresocks"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/model"
)

func GenerateWireguardConfig(i *model.Identity) wiresocks.Configuration {
	priv, _ := wiresocks.EncodeBase64ToHex(i.PrivateKey)
	pub, _ := wiresocks.EncodeBase64ToHex(i.Config.Peers[0].PublicKey)

	var dnsAddrs []netip.Addr
	for _, dns := range []string{"1.1.1.1", "1.0.0.1", "2606:4700:4700::1112", "2606:4700:4700::1112"} {
		addr := netip.MustParseAddr(dns)
		dnsAddrs = append(dnsAddrs, addr)
	}

	return wiresocks.Configuration{
		Interface: &wiresocks.InterfaceConfig{
			PrivateKey: priv,
			Addresses: []netip.Prefix{
				wiresocks.MustParsePrefixOrAddr(i.Config.Interface.Addresses.V4),
				wiresocks.MustParsePrefixOrAddr(i.Config.Interface.Addresses.V6),
			},
			MTU: 1280,
			DNS: dnsAddrs,
		},
		Peers: []wiresocks.PeerConfig{{
			PublicKey:    pub,
			PreSharedKey: "0000000000000000000000000000000000000000000000000000000000000000",
			AllowedIPs: []netip.Prefix{
				netip.MustParsePrefix("0.0.0.0/0"),
				netip.MustParsePrefix("::/0"),
			},
			KeepAlive: 25,
			Endpoint:  i.Config.Peers[0].Endpoint.Host,
		}},
	}
}

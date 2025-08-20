package ping

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/flynn/noise"
	"go.uber.org/zap"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/curve25519"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/network"
	"github.com/shahradelahi/cloudflare-warp/ipscanner/model"
	"github.com/shahradelahi/cloudflare-warp/log"
	"github.com/shahradelahi/cloudflare-warp/utils"
)

const (
	minRandomPackets       = 20
	maxRandomPackets       = 50
	minRandomPacketSize    = 10  // Relative to header size
	maxRandomPacketSize    = 120 // Relative to header size
	minRandomPacketDelayMs = 80
	maxRandomPacketDelayMs = 150
)

type WarpPingResult struct {
	AddrPort netip.AddrPort
	RTT      time.Duration
	Err      error
}

func (h *WarpPingResult) Result() statute.IPInfo {
	return statute.IPInfo{AddrPort: h.AddrPort, RTT: h.RTT, CreatedAt: time.Now()}
}

func (h *WarpPingResult) Error() error {
	return h.Err
}

func (h *WarpPingResult) String() string {
	if h.Err != nil {
		return fmt.Sprintf("%s", h.Err)
	} else {
		return fmt.Sprintf("%s: protocol=%s, time=%d ms", h.AddrPort, "warp", h.RTT)
	}
}

type WarpPing struct {
	PrivateKey    string
	PeerPublicKey string
	PresharedKey  string
	IP            netip.Addr
	opts          *statute.ScannerOptions
}

func (h *WarpPing) Ping() statute.IPingResult {
	return h.PingContext(context.Background())
}

func (h *WarpPing) PingContext(ctx context.Context) statute.IPingResult {
	ports := network.ScannerPorts()
	results := make(chan statute.IPingResult, len(ports))
	var wg sync.WaitGroup

	for _, port := range ports {
		wg.Add(1)
		addr := netip.AddrPortFrom(h.IP, port)
		go func(addr netip.AddrPort) {
			defer wg.Done()
			log.Debugf("Attempting to ping WARP endpoint: %s", addr.String())
			rtt, err := initiateHandshake(
				ctx,
				addr,
				h.PrivateKey,
				h.PeerPublicKey,
				h.PresharedKey,
				true, // Obfuscation
			)
			if err == nil {
				log.Debugf("Successfully pinged WARP endpoint %s, RTT: %s", addr.String(), rtt.String())
				results <- &WarpPingResult{AddrPort: addr, RTT: rtt, Err: nil}
			} else {
				log.Debugw("Failed to ping WARP endpoint", zap.String("address", addr.String()), zap.Error(err))
				if h.opts != nil && h.opts.EventsHandler != nil {
					h.opts.EventsHandler.IncrementFailure(addr.String())
				}
				results <- h.errorResult(err)
			}
		}(addr)
	}

	wg.Wait()
	close(results)

	var lastErr error
	for res := range results {
		if res.Error() == nil {
			return res
		}
		lastErr = res.Error()
	}

	return h.errorResult(lastErr)
}

func (h *WarpPing) errorResult(err error) *WarpPingResult {
	r := &WarpPingResult{}
	r.Err = err
	return r
}

func staticKeypair(privateKeyBase64 string) (noise.DHKey, error) {
	privateKey, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return noise.DHKey{}, err
	}

	var pubkey, privkey [32]byte
	copy(privkey[:], privateKey)
	curve25519.ScalarBaseMult(&pubkey, &privkey)

	return noise.DHKey{
		Private: privateKey,
		Public:  pubkey[:],
	}, nil
}

func ephemeralKeypair() (noise.DHKey, error) {
	// Generate an ephemeral private key
	ephemeralPrivateKey := make([]byte, 32)
	if _, err := rand.Read(ephemeralPrivateKey); err != nil {
		return noise.DHKey{}, err
	}

	// Derive the corresponding ephemeral public key
	ephemeralPublicKey, err := curve25519.X25519(ephemeralPrivateKey, curve25519.Basepoint)
	if err != nil {
		return noise.DHKey{}, err
	}

	return noise.DHKey{
		Private: ephemeralPrivateKey,
		Public:  ephemeralPublicKey,
	}, nil
}

func generateObfuscationHeader() ([]byte, error) {
	clist := []byte{0xDC, 0xDE, 0xD3, 0xD9, 0xD0, 0xEC, 0xEE, 0xE3}
	firstByteIndex, err := utils.RandomInt(0, uint64(len(clist)-1))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random byte for header: %w", err)
	}
	header := []byte{
		clist[firstByteIndex],
		0x00, 0x00, 0x00, 0x01, 0x08,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x44, 0xD0,
	}
	_, err = rand.Read(header[6:14])
	if err != nil {
		return nil, fmt.Errorf("failed to generate random part of header: %w", err)
	}
	return header, nil
}

func sendRandomPackets(ctx context.Context, conn net.Conn, obfuscation bool) error {
	if !obfuscation {
		return nil
	}

	header, err := generateObfuscationHeader()
	if err != nil {
		return fmt.Errorf("failed to generate obfuscation header: %w", err)
	}
	if header == nil {
		return nil
	}

	numPackets, err := utils.RandomInt(minRandomPackets, maxRandomPackets)
	if err != nil {
		return fmt.Errorf("failed to generate random packet count: %w", err)
	}

	maxPacketSize := uint64(len(header)) + maxRandomPacketSize
	randomPacket := make([]byte, maxPacketSize)

	for i := uint64(0); i < numPackets; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			packetSize, err := utils.RandomInt(uint64(len(header))+minRandomPacketSize, maxPacketSize)
			if err != nil {
				return fmt.Errorf("failed to generate random packet size: %w", err)
			}

			// Fill random payload
			if packetSize > uint64(len(header)) {
				_, err = rand.Read(randomPacket[len(header):packetSize])
				if err != nil {
					return fmt.Errorf("failed to generate random packet payload: %w", err)
				}
			}

			// Copy header
			copy(randomPacket[:len(header)], header)

			_, err = conn.Write(randomPacket[:packetSize])
			if err != nil {
				return fmt.Errorf("error sending random packet: %w", err)
			}

			delay, err := utils.RandomInt(minRandomPacketDelayMs, maxRandomPacketDelayMs)
			if err != nil {
				log.Warnw("Failed to generate random delay", zap.Error(err))
			} else {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}
	}
	return nil
}

func initiateHandshake(ctx context.Context, serverAddr netip.AddrPort, privateKeyBase64, peerPublicKeyBase64, presharedKeyBase64 string, obfuscation bool) (time.Duration, error) {
	staticKeyPair, err := staticKeypair(privateKeyBase64)
	if err != nil {
		return 0, err
	}

	peerPublicKey, err := base64.StdEncoding.DecodeString(peerPublicKeyBase64)
	if err != nil {
		return 0, err
	}

	presharedKey, err := base64.StdEncoding.DecodeString(presharedKeyBase64)
	if err != nil {
		return 0, err
	}

	if presharedKeyBase64 == "" {
		presharedKey = make([]byte, 32)
	}

	ephemeral, err := ephemeralKeypair()
	if err != nil {
		return 0, err
	}

	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashBLAKE2s)
	hs, err := noise.NewHandshakeState(noise.Config{
		CipherSuite:           cs,
		Pattern:               noise.HandshakeIK,
		Initiator:             true,
		StaticKeypair:         staticKeyPair,
		PeerStatic:            peerPublicKey,
		Prologue:              []byte("WireGuard v1 zx2c4 Jason@zx2c4.com"),
		PresharedKey:          presharedKey,
		PresharedKeyPlacement: 2,
		EphemeralKeypair:      ephemeral,
		Random:                rand.Reader,
	})
	if err != nil {
		return 0, err
	}

	// Prepare handshake initiation packet

	// TAI64N timestamp calculation
	now := time.Now().UTC()
	epochOffset := int64(4611686018427387914) // TAI offset from Unix epoch

	tai64nTimestampBuf := make([]byte, 0, 16)
	tai64nTimestampBuf = binary.BigEndian.AppendUint64(tai64nTimestampBuf, uint64(epochOffset+now.Unix()))
	tai64nTimestampBuf = binary.BigEndian.AppendUint32(tai64nTimestampBuf, uint32(now.Nanosecond()))
	msg, _, _, err := hs.WriteMessage(nil, tai64nTimestampBuf)
	if err != nil {
		return 0, err
	}

	initiationPacket := new(bytes.Buffer)
	binary.Write(initiationPacket, binary.BigEndian, []byte{0x01, 0x00, 0x00, 0x00})
	binary.Write(initiationPacket, binary.BigEndian, utils.Uint32ToBytes(28))
	binary.Write(initiationPacket, binary.BigEndian, msg)

	macKey := blake2s.Sum256(append([]byte("mac1----"), peerPublicKey...))
	hasher, err := blake2s.New128(macKey[:]) // using macKey as the key
	if err != nil {
		return 0, err
	}
	_, err = hasher.Write(initiationPacket.Bytes())
	if err != nil {
		return 0, err
	}
	initiationPacketMAC := hasher.Sum(nil)

	// Append the MAC and 16 null bytes to the initiation packet
	binary.Write(initiationPacket, binary.BigEndian, initiationPacketMAC[:16])
	binary.Write(initiationPacket, binary.BigEndian, [16]byte{})

	conn, err := net.Dial("udp", serverAddr.String())
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	if err := sendRandomPackets(ctx, conn, obfuscation); err != nil {
		return 0, err
	}

	_, err = initiationPacket.WriteTo(conn)
	if err != nil {
		return 0, err
	}
	t0 := time.Now()

	response := make([]byte, 92)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	i, err := conn.Read(response)
	if err != nil {
		return 0, err
	}
	rtt := time.Since(t0)

	if i < 60 {
		return 0, fmt.Errorf("invalid handshake response length %d bytes", i)
	}

	// Check the response type
	if response[0] != 2 { // 2 is the message type for response
		return 0, errors.New("invalid response type")
	}

	// Extract sender and receiver index from the response
	// peer index
	_ = binary.LittleEndian.Uint32(response[4:8])
	// our index(we set it to 28)
	ourIndex := binary.LittleEndian.Uint32(response[8:12])
	if ourIndex != 28 { // Check if the response corresponds to our sender index
		return 0, errors.New("invalid sender index in response")
	}

	payload, _, _, err := hs.ReadMessage(nil, response[12:60])
	if err != nil {
		return 0, err
	}

	// Check if the payload is empty (as expected in WireGuard handshake)
	if len(payload) != 0 {
		return 0, errors.New("unexpected payload in response")
	}

	return rtt, nil
}

func NewWarpPing(ip netip.Addr, opts *statute.ScannerOptions) *WarpPing {
	return &WarpPing{
		PrivateKey:    opts.WarpPrivateKey,
		PeerPublicKey: opts.WarpPeerPublicKey,
		PresharedKey:  opts.WarpPresharedKey,
		IP:            ip,
		opts:          opts,
	}
}

var (
	_ statute.IPing       = (*WarpPing)(nil)
	_ statute.IPingResult = (*WarpPingResult)(nil)
)

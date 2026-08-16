// Package sdk provides an embeddable API for running one or more Cloudflare
// WARP-backed proxy servers in a Go application.
package sdk

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/shahradelahi/cloudflare-warp/cloudflare"
	"github.com/shahradelahi/cloudflare-warp/cloudflare/model"
	"github.com/shahradelahi/cloudflare-warp/core/datadir"
)

// ClientConfig controls WARP identity loading.
type ClientConfig struct {
	// DataDir contains reg.json and conf.json. Missing identity files are
	// created automatically. The default is ~/.cloudflare-warp.
	DataDir string

	// Identity skips filesystem and Cloudflare API identity loading when set.
	// This is useful when an application manages WARP credentials itself.
	Identity *model.Identity
}

// Client owns a WARP identity and creates independent proxy instances.
// A Client is safe for concurrent use.
type Client struct {
	identity *model.Identity
	dataDir  string
}

// datadirMu protects the legacy global data directory while an identity is
// loaded or created. Proxies use Client.identity after NewClient returns.
var datadirMu sync.Mutex

// NewClient loads or creates one WARP identity. Reuse the returned Client for
// all proxies in an application to avoid duplicate registrations.
func NewClient(config ClientConfig) (*Client, error) {
	if config.Identity != nil {
		if err := validateIdentity(config.Identity); err != nil {
			return nil, err
		}
		return &Client{identity: config.Identity, dataDir: config.DataDir}, nil
	}

	dir := datadir.GetDataDirOrPath(config.DataDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create WARP data directory: %w", err)
	}

	datadirMu.Lock()
	defer datadirMu.Unlock()

	datadir.SetDataDir(dir)
	identity, err := cloudflare.LoadOrCreateIdentity()
	if err != nil {
		return nil, fmt.Errorf("load or create WARP identity: %w", err)
	}
	if err := validateIdentity(identity); err != nil {
		return nil, err
	}

	return &Client{identity: identity, dataDir: dir}, nil
}

// GenerateIdentity loads an existing WARP identity from dataDir or generates
// and saves a new one there when identity files do not exist.
func GenerateIdentity(dataDir string) (*Client, error) {
	if dataDir == "" {
		return nil, errors.New("WARP identity data directory is required")
	}
	return NewClient(ClientConfig{DataDir: dataDir})
}

// DataDir returns the identity directory used by this client. It is empty when
// ClientConfig.Identity was supplied without a directory.
func (c *Client) DataDir() string {
	if c == nil {
		return ""
	}
	return c.dataDir
}

func validateIdentity(identity *model.Identity) error {
	if identity == nil {
		return errors.New("WARP identity is nil")
	}
	if identity.PrivateKey == "" {
		return errors.New("WARP identity has no private key")
	}
	if len(identity.Config.Peers) == 0 {
		return errors.New("WARP identity has no peers")
	}
	if identity.Config.Peers[0].PublicKey == "" {
		return errors.New("WARP identity peer has no public key")
	}
	return nil
}

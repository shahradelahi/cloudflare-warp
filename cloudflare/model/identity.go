package model

type IdentityConfigPeerEndpoint struct {
	V4    string   `json:"v4"`
	V6    string   `json:"v6"`
	Host  string   `json:"host"`
	Ports []uint16 `json:"ports"`
}

type IdentityConfigPeer struct {
	PublicKey string                     `json:"public_key"`
	Endpoint  IdentityConfigPeerEndpoint `json:"endpoint"`
}

type IdentityConfigInterfaceAddresses struct {
	V4 string `json:"v4"`
	V6 string `json:"v6"`
}

type IdentityConfigInterface struct {
	Addresses IdentityConfigInterfaceAddresses `json:"addresses"`
}
type IdentityConfigServices struct {
	HTTPProxy string `json:"http_proxy"`
}

type IdentityConfig struct {
	Peers     []IdentityConfigPeer    `json:"peers"`
	Interface IdentityConfigInterface `json:"interface"`
	Services  IdentityConfigServices  `json:"services"`
	ClientID  string                  `json:"client_id"`
}

type Identity struct {
	Version         string          `json:"version,omitempty"`
	PrivateKey      string          `json:"private_key"`
	Key             string          `json:"key"`
	Account         IdentityAccount `json:"account"`
	Place           int64           `json:"place"`
	FCMToken        string          `json:"fcm_token"`
	Name            string          `json:"name"`
	TOS             string          `json:"tos"`
	Locale          string          `json:"locale"`
	InstallID       string          `json:"install_id"`
	WarpEnabled     bool            `json:"warp_enabled"`
	Type            string          `json:"type"`
	Model           string          `json:"model"`
	Config          IdentityConfig  `json:"config"`
	Token           string          `json:"token"`
	Enabled         bool            `json:"enabled"`
	ID              string          `json:"id"`
	Created         string          `json:"created"`
	Updated         string          `json:"updated"`
	WaitlistEnabled bool            `json:"waitlist_enabled"`
}

type IdentityDevice struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Model     string `json:"model"`
	Created   string `json:"created"`
	Activated string `json:"updated"`
	Active    bool   `json:"active"`
	Role      string `json:"role"`
}

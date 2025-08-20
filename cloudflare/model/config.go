package model

// RegFile represents the structure of reg.json
type RegFile struct {
	RegistrationID string `json:"registration_id"`
	Token          string `json:"token"`
	PrivateKey     string `json:"private_key"` // This will be the WireGuard private key
}

// ConfFile represents the structure of conf.json
type ConfFile struct {
	Account IdentityAccount `json:"account"`
	Config  IdentityConfig  `json:"config"`
}

// SettingsFile represents the structure of settings.json
type SettingsFile struct {
	OperationMode string `json:"operation_mode"`
}

package model

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/shahradelahi/cloudflare-warp/core/datadir"
)

func GetRegPath() string {
	return filepath.Join(datadir.GetDataDir(), "reg.json")
}

func GetConfPath() string {
	return filepath.Join(datadir.GetDataDir(), "conf.json")
}

func (a *Identity) SaveIdentity() error {

	regPath := GetRegPath()
	confPath := GetConfPath()

	// Save reg.json
	regFileContent := RegFile{
		RegistrationID: a.ID,
		Token:          a.Token,
		PrivateKey:     a.PrivateKey,
	}
	regData, err := json.MarshalIndent(regFileContent, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(regPath, regData, 0600); err != nil {
		return err
	}

	// Save conf.json
	confFileContent := ConfFile{
		Account: a.Account,
		Config:  a.Config,
	}
	confData, err := json.MarshalIndent(confFileContent, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(confPath, confData, 0600)
}

package cloudflare

import (
	"encoding/json"
	"errors"
	"os"

	"go.uber.org/zap"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/crypto"
	"github.com/shahradelahi/cloudflare-warp/cloudflare/model"
	"github.com/shahradelahi/cloudflare-warp/log"
)

func CreateOrUpdateIdentity(license string) (*model.Identity, error) {
	warpAPI := NewWarpAPI()

	identity, err := LoadIdentity()
	if err != nil {
		log.Warnw("Failed to load existing WARP identity; attempting to create a new one", zap.Error(err))

		log.Info("Initiating creation of a new WARP identity...")
		newIdentity, err := CreateIdentity(warpAPI, license)
		if err != nil {
			return nil, err
		}
		return &newIdentity, nil
	}

	if license != "" && identity.Account.License != license {
		log.Info("Attempting to update WARP account license key...")
		_, err := warpAPI.UpdateAccount(identity.Token, identity.ID, license)
		if err != nil {
			return nil, err
		}

		iAcc, err := warpAPI.GetAccount(identity.Token, identity.ID)
		if err != nil {
			return nil, err
		}
		identity.Account = iAcc
	}

	return &identity, nil
}

func LoadOrCreateIdentity() (*model.Identity, error) {
	identity, err := CreateOrUpdateIdentity("")
	if err != nil {
		return nil, err
	}

	if err = identity.SaveIdentity(); err != nil {
		return nil, err
	}

	log.Debug("Successfully loaded WARP identity.")
	return identity, nil
}

func LoadIdentity() (model.Identity, error) {
	regPath := model.GetRegPath()
	confPath := model.GetConfPath()

	if _, err := os.Stat(regPath); os.IsNotExist(err) {
		return model.Identity{}, err
	}
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		return model.Identity{}, err
	}

	regBytes, err := os.ReadFile(regPath)
	if err != nil {
		return model.Identity{}, err
	}
	confBytes, err := os.ReadFile(confPath)
	if err != nil {
		return model.Identity{}, err
	}

	var regFile model.RegFile
	if err := json.Unmarshal(regBytes, &regFile); err != nil {
		return model.Identity{}, err
	}

	var confFile model.ConfFile
	if err := json.Unmarshal(confBytes, &confFile); err != nil {
		return model.Identity{}, err
	}

	identity := model.Identity{
		ID:         regFile.RegistrationID,
		Token:      regFile.Token,
		PrivateKey: regFile.PrivateKey,
		Account:    confFile.Account,
		Config:     confFile.Config,
		Version:    "v2", // new version
	}

	if len(identity.Config.Peers) < 1 {
		return model.Identity{}, errors.New("identity contains 0 peers")
	}

	return identity, nil
}

func CreateIdentity(warpAPI *WarpAPI, license string) (model.Identity, error) {
	priv, err := crypto.GeneratePrivateKey()
	if err != nil {
		return model.Identity{}, err
	}

	privateKey, publicKey := priv.String(), priv.PublicKey().String()

	i, err := warpAPI.Register(publicKey)
	if err != nil {
		return model.Identity{}, err
	}

	if license != "" {
		log.Info("Attempting to update WARP account license key...")
		_, err := warpAPI.UpdateAccount(i.Token, i.ID, license)
		if err != nil {
			return model.Identity{}, err
		}

		ac, err := warpAPI.GetAccount(i.Token, i.ID)
		if err != nil {
			return model.Identity{}, err
		}
		i.Account = ac
	}

	i.PrivateKey = privateKey
	i.Version = "v2"

	return i, nil
}

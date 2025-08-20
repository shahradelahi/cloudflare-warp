package cloudflare

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/model"
	"github.com/shahradelahi/cloudflare-warp/core/datadir"
)

func TestLoadOrCreateIdentity(t *testing.T) {
	// Skip this test in CI environments as it requires a live connection
	// and might be flaky.
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test that requires live API connection in CI environment")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-datadir")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	datadir.SetDataDir(tempDir)

	// Test creating a new identity with the provided license key
	// Using a real license key makes this an integration test.
	licenseKey := "5m3o6Qq4-495D2Kpk-egrG8326"
	identity, err := CreateOrUpdateIdentity(licenseKey)
	assert.NoError(t, err)
	assert.NotNil(t, identity)
	assert.NotEmpty(t, identity.ID, "Identity ID should not be empty")
	assert.NotEmpty(t, identity.Token, "Identity Token should not be empty")
	assert.Equal(t, licenseKey, identity.Account.License)
	assert.True(t, identity.Account.WarpPlus, "WarpPlus should be enabled with a valid license")

	// Verify that the configuration files were created
	regPath := model.GetRegPath()
	confPath := model.GetConfPath()

	_, err = os.Stat(regPath)
	assert.NoError(t, err, "reg.json should be created")
	_, err = os.Stat(confPath)
	assert.NoError(t, err, "conf.json should be created")

	// Test loading the created identity from the files
	loadedIdentity, err := LoadIdentity()
	assert.NoError(t, err)
	assert.NotNil(t, loadedIdentity)
	assert.Equal(t, identity.ID, loadedIdentity.ID)
	assert.Equal(t, identity.Token, loadedIdentity.Token)
	assert.Equal(t, identity.Account.License, loadedIdentity.Account.License)
}

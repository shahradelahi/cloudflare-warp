package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privateKey, err := GeneratePrivateKey()
	assert.NoError(t, err, "GeneratePrivateKey should not return an error")
	assert.NotNil(t, privateKey, "Generated private key should not be nil")

	// Check key length
	assert.Equal(t, KeyLen, len(privateKey), "Private key should have the correct length")

	// Check if the key is properly clamped, according to Curve25519 spec.
	// https://cr.yp.to/ecdh.html
	assert.Equal(t, byte(0), privateKey[0]&0b111, "The 3 least significant bits of the first byte should be 0")
	assert.Equal(t, byte(0), privateKey[31]&0b10000000, "The most significant bit of the last byte should be 0")
	assert.NotEqual(t, byte(0), privateKey[31]&0b01000000, "The second most significant bit of the last byte should be 1")
}

func TestPublicKey(t *testing.T) {
	privateKey, err := GeneratePrivateKey()
	assert.NoError(t, err)

	publicKey := privateKey.PublicKey()
	assert.NotNil(t, publicKey, "Generated public key should not be nil")

	// Check key length
	assert.Equal(t, KeyLen, len(publicKey), "Public key should have the correct length")

	// A public key should not be equal to its private key
	assert.NotEqual(t, privateKey, publicKey, "Public key should be different from the private key")
}

func TestNewKey(t *testing.T) {
	// Test with a valid key
	validBytes := make([]byte, KeyLen)
	_, err := NewKey(validBytes)
	assert.NoError(t, err, "NewKey should not return an error for a valid key")

	// Test with an invalid key (wrong length)
	invalidBytes := make([]byte, KeyLen-1)
	_, err = NewKey(invalidBytes)
	assert.Error(t, err, "NewKey should return an error for a key with incorrect length")
}

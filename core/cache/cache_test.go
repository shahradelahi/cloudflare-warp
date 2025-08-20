package cache

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/shahradelahi/cloudflare-warp/core/datadir"
)

func setupTestCache(t *testing.T) (*Cache, func()) {
	tempDir, err := os.MkdirTemp("", "cache-test")
	assert.NoError(t, err)
	datadir.SetDataDir(tempDir)

	// Create a new cache for each test to ensure isolation
	c := &Cache{
		Endpoints: make([]Endpoint, 0),
	}
	// SetDataDir the singleton instance for any code that might still rely on it.
	instance = c

	return c, func() {
		os.RemoveAll(tempDir)
	}
}

func TestSaveEndpoint(t *testing.T) {
	c, cleanup := setupTestCache(t)
	defer cleanup()

	c.SaveEndpoint("1.1.1.1:2408", 100*time.Millisecond)
	assert.Equal(t, 1, len(c.Endpoints))
	assert.Equal(t, "1.1.1.1:2408", c.Endpoints[0].Address)
	assert.Equal(t, 100*time.Millisecond, c.Endpoints[0].RTT)
	assert.Equal(t, 0, c.Endpoints[0].Failures)

	// Test updating an existing endpoint
	c.SaveEndpoint("1.1.1.1:2408", 50*time.Millisecond)
	assert.Equal(t, 1, len(c.Endpoints))
	assert.Equal(t, 50*time.Millisecond, c.Endpoints[0].RTT)
}

func TestRecordFailureAndSuccess(t *testing.T) {
	c, cleanup := setupTestCache(t)
	defer cleanup()

	c.SaveEndpoint("1.1.1.1:2408", 100*time.Millisecond)

	// Record failures
	c.RecordFailure("1.1.1.1:2408")
	assert.Equal(t, 1, c.Endpoints[0].Failures)

	c.RecordFailure("1.1.1.1:2408")
	assert.Equal(t, 2, c.Endpoints[0].Failures)

	// Record success, should reset failures
	c.RecordSuccess("1.1.1.1:2408")
	assert.Equal(t, 0, c.Endpoints[0].Failures)
}

func TestEndpointRemoval(t *testing.T) {
	c, cleanup := setupTestCache(t)
	defer cleanup()

	c.SaveEndpoint("1.1.1.1:2408", 100*time.Millisecond)

	// Fail endpoint until it's removed
	for i := 0; i < maxFailures; i++ {
		c.RecordFailure("1.1.1.1:2408")
	}

	assert.Equal(t, 0, len(c.Endpoints), "Endpoint should be removed after maxFailures")
}

func TestGetEndpoints(t *testing.T) {
	c, cleanup := setupTestCache(t)
	defer cleanup()

	c.SaveEndpoint("1.1.1.1:2408", 100*time.Millisecond)
	c.SaveEndpoint("2.2.2.2:2408", 200*time.Millisecond)
	c.SaveEndpoint("3.3.3.3:2408", 50*time.Millisecond)

	c.RecordFailure("2.2.2.2:2408")

	// Test GetBestEndpoint
	best, err := c.GetBestEndpoint()
	assert.NoError(t, err)
	assert.Equal(t, "3.3.3.3:2408", best.Address, "Best endpoint should have the lowest RTT")

	// Test GetAllEndpoints
	all := c.GetAllEndpoints()
	assert.Equal(t, 3, len(all))
	assert.Equal(t, "1.1.1.1:2408", all[0].Address, "Endpoints should be sorted by failures")
	assert.Equal(t, "3.3.3.3:2408", all[1].Address)
	assert.Equal(t, "2.2.2.2:2408", all[2].Address)

	// Test GetRandomEndpoint
	random, err := c.GetRandomEndpoint()
	assert.NoError(t, err)
	assert.NotNil(t, random)
}

func TestLoadAndSaveCache(t *testing.T) {
	c, cleanup := setupTestCache(t)
	defer cleanup()

	c.SaveEndpoint("1.1.1.1:2408", 100*time.Millisecond)
	c.Endpoints[0].Timestamp = time.Time{} // Zero out timestamp for consistent comparison
	err := c.SaveCache()
	assert.NoError(t, err)

	// Create a new cache instance to load into
	c2 := &Cache{
		Endpoints: make([]Endpoint, 0),
	}
	err = c2.LoadCache()
	assert.NoError(t, err)

	assert.Equal(t, 1, len(c2.Endpoints))
	c2.Endpoints[0].Timestamp = time.Time{} // Zero out timestamp for consistent comparison
	assert.Equal(t, c.Endpoints, c2.Endpoints)
}

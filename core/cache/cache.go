package cache

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/shahradelahi/cloudflare-warp/core/datadir"
)

const (
	maxFailures = 3
)

var (
	instance *Cache
	once     sync.Once
)

// Endpoint represents a WARP endpoint with its details.
type Endpoint struct {
	Address   string        `json:"address"`
	RTT       time.Duration `json:"rtt"`
	Timestamp time.Time     `json:"timestamp"`
	Failures  int           `json:"failures"`
}

// Cache stores the cached endpoints.
type Cache struct {
	Endpoints []Endpoint `json:"endpoints"`
	mutex     sync.Mutex
}

// NewCache creates a new Cache instance.
func NewCache() *Cache {
	once.Do(func() {
		instance = &Cache{
			Endpoints: make([]Endpoint, 0),
		}
	})
	return instance
}

// SaveEndpoint saves a new endpoint to the cache or updates an existing one.
func (c *Cache) SaveEndpoint(address string, rtt time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, endpoint := range c.Endpoints {
		if endpoint.Address == address {
			c.Endpoints[i].RTT = rtt
			c.Endpoints[i].Timestamp = time.Now()
			c.Endpoints[i].Failures = 0
			return
		}
	}

	c.Endpoints = append(c.Endpoints, Endpoint{
		Address:   address,
		RTT:       rtt,
		Timestamp: time.Now(),
		Failures:  0,
	})
}

// GetBestEndpoint retrieves the endpoint with the lowest RTT that has not failed more than maxFailures.
func (c *Cache) GetBestEndpoint() (*Endpoint, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var availableEndpoints []Endpoint
	for _, endpoint := range c.Endpoints {
		if endpoint.Failures < maxFailures {
			availableEndpoints = append(availableEndpoints, endpoint)
		}
	}

	if len(availableEndpoints) == 0 {
		return nil, fmt.Errorf("no available endpoints in the cache")
	}

	sort.Slice(availableEndpoints, func(i, j int) bool {
		return availableEndpoints[i].RTT < availableEndpoints[j].RTT
	})

	return &availableEndpoints[0], nil
}

// GetRandomEndpoint retrieves a random endpoint that has not failed more than maxFailures.
func (c *Cache) GetRandomEndpoint() (*Endpoint, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var availableEndpoints []Endpoint
	for _, endpoint := range c.Endpoints {
		if endpoint.Failures < maxFailures {
			availableEndpoints = append(availableEndpoints, endpoint)
		}
	}

	if len(availableEndpoints) == 0 {
		return nil, fmt.Errorf("no available endpoints in the cache")
	}

	return &availableEndpoints[rand.Intn(len(availableEndpoints))], nil
}

// GetAllEndpoints retrieves all endpoints that have not failed more than maxFailures, sorted by failures.
func (c *Cache) GetAllEndpoints() []Endpoint {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var availableEndpoints []Endpoint
	for _, endpoint := range c.Endpoints {
		if endpoint.Failures < maxFailures {
			availableEndpoints = append(availableEndpoints, endpoint)
		}
	}

	sort.Slice(availableEndpoints, func(i, j int) bool {
		return availableEndpoints[i].Failures < availableEndpoints[j].Failures
	})

	return availableEndpoints
}

// RecordFailure increments the failure count for a given endpoint and removes it if it exceeds the maxFailures.
func (c *Cache) RecordFailure(address string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, endpoint := range c.Endpoints {
		if endpoint.Address == address {
			c.Endpoints[i].Failures++
			if c.Endpoints[i].Failures >= maxFailures {
				c.Endpoints = append(c.Endpoints[:i], c.Endpoints[i+1:]...)
			}
			return
		}
	}
}

// RecordSuccess resets the failure count for a given endpoint.
func (c *Cache) RecordSuccess(address string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, endpoint := range c.Endpoints {
		if endpoint.Address == address {
			c.Endpoints[i].Failures = 0
			return
		}
	}
}

// LoadCache loads the cache from a file.
func (c *Cache) LoadCache() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	dir := datadir.GetDataDir()
	if dir == "" {
		return fmt.Errorf("data directory not set")
	}
	filePath := filepath.Join(dir, "endpoints.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Cache file doesn't exist yet, which is fine.
		}
		return err
	}

	return json.Unmarshal(data, &c.Endpoints)
}

// SaveCache saves the cache to a file.
func (c *Cache) SaveCache() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	dir := datadir.GetDataDir()
	if dir == "" {
		return fmt.Errorf("data directory not set")
	}
	filePath := filepath.Join(dir, "endpoints.json")

	data, err := json.MarshalIndent(c.Endpoints, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func (c *Cache) GetRandomEndpoints(count int) ([]string, error) {
	var endpoints []string
	for i := 0; i < count; i++ {
		endpoint, err := c.GetRandomEndpoint()
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint.Address)
	}
	return endpoints, nil
}

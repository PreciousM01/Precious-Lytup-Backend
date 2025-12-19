package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type CacheEntry struct {
	Data      any
	ExpiresAt time.Time
}

type WeatherCache struct {
	Store map[string]CacheEntry
	Mutex sync.RWMutex
	TTL   time.Duration
}

var GlobalWeatherCache = &WeatherCache{
	Store: make(map[string]CacheEntry),
	TTL:   30 * time.Minute, 
}

// Generate cache key from city and dates
func (c *WeatherCache) GenerateKey(city, startDate, endDate string) string {
	raw := fmt.Sprintf("%s:%s:%s", city, startDate, endDate)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// Get cached data if it exists and hasn't expired
func (c *WeatherCache) Get(key string) (any, bool) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	entry, exists := c.Store[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Data, true
}

// Store data in cache with TTL
func (c *WeatherCache) Set(key string, data any) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.Store[key] = CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.TTL),
	}
}

func (c *WeatherCache) SetWithTTL(key string, data any, ttl time.Duration) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	c.Store[key] = CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl), // Use custom TTL
	}
}

// Clean up expired entries (run periodically)
func (c *WeatherCache) CleanExpired() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	now := time.Now()
	for key, entry := range c.Store {
		if now.After(entry.ExpiresAt) {
			delete(c.Store, key)
		}
	}
}

// Start background cleanup goroutine
func (c *WeatherCache) StartCleanup() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			c.CleanExpired()
			Output_Logg("OUT", "Cache", fmt.Sprintf("Cleaned expired cache entries. Current size: %d", len(c.Store)))
		}
	}()
}
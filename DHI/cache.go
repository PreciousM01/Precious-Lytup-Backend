package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type CacheEntry struct {
	Data      any       `json:"data"`
	ExpiresAt time.Time `json:"expires_at"`
	StoredAt  time.Time `json:"stored_at"`
}

type WeatherCache struct {
	Store    map[string]CacheEntry
	Mutex    sync.RWMutex
	TTL      time.Duration
	FilePath string // Path to persistent cache file
}

var GlobalWeatherCache = &WeatherCache{
	Store:    make(map[string]CacheEntry),
	TTL:      30 * time.Minute,
	FilePath: "weather_cache.json",
}

// 1. Generate cache key from city and dates
func (c *WeatherCache) GenerateKey(city, startDate, endDate string) string {
	raw := fmt.Sprintf("%s:%s:%s", city, startDate, endDate)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// 2. Get cached data if it exists and hasn't expired
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

// 3. Get stale cache - returns data even if expired (for fallback)
func (c *WeatherCache) GetStale(key string, maxAge time.Duration) (any, bool, time.Duration) {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	entry, exists := c.Store[key]
	if !exists {
		return nil, false, 0
	}

	age := time.Since(entry.StoredAt)

	if age > maxAge {
		return nil, false, 0
	}

	return entry.Data, true, age
}

// 4. Store data in cache with TTL
func (c *WeatherCache) Set(key string, data any) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	now := time.Now()
	c.Store[key] = CacheEntry{
		Data:      data,
		ExpiresAt: now.Add(c.TTL),
		StoredAt:  now,
	}
}

// 5. Store data with custom TTL
func (c *WeatherCache) SetWithTTL(key string, data any, ttl time.Duration) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	now := time.Now()
	c.Store[key] = CacheEntry{
		Data:      data,
		ExpiresAt: now.Add(ttl),
		StoredAt:  now,
	}
}

// 6. Clean up expired entries (run periodically)
func (c *WeatherCache) CleanExpired() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	now := time.Now()
	maxStaleAge := 24 * time.Hour

	for key, entry := range c.Store {
		age := now.Sub(entry.StoredAt)
		if age > maxStaleAge {
			delete(c.Store, key)
		}
	}
}

// 7. Start background cleanup goroutine
func (c *WeatherCache) StartCleanup() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			c.CleanExpired()
			Output_Logg("OUT", "Cache", fmt.Sprintf("Cleaned expired entries. Current size: %d", len(c.Store)))
		}
	}()
}

// 8. Save cache to disk
func (c *WeatherCache) Save() error {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	// Create a serializable version of the cache
	data, err := json.MarshalIndent(c.Store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write to file
	if err := os.WriteFile(c.FilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	Output_Logg("OUT", "Cache", fmt.Sprintf("Saved %d entries to %s", len(c.Store), c.FilePath))
	return nil
}

// 9. Load cache from disk
func (c *WeatherCache) Load() error {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(c.FilePath); os.IsNotExist(err) {
		Output_Logg("OUT", "Cache", "No existing cache file found. Starting with empty cache.")
		return nil
	}

	// Read file
	data, err := os.ReadFile(c.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	// Unmarshal into temporary map
	var tempStore map[string]CacheEntry
	if err := json.Unmarshal(data, &tempStore); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	// Filter out expired entries during load
	now := time.Now()
	maxStaleAge := 24 * time.Hour
	validCount := 0

	for key, entry := range tempStore {
		age := now.Sub(entry.StoredAt)
		if age <= maxStaleAge {
			c.Store[key] = entry
			validCount++
		}
	}

	Output_Logg("OUT", "Cache", fmt.Sprintf("Loaded %d valid entries from %s (%d expired entries discarded)", 
		validCount, c.FilePath, len(tempStore)-validCount))

	return nil
}

// 10. Get cache statistics
func (c *WeatherCache) GetStats() map[string]int {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	now := time.Now()
	stats := map[string]int{
		"total":   0,
		"fresh":   0,
		"stale":   0,
		"expired": 0,
	}

	for _, entry := range c.Store {
		stats["total"]++

		age := now.Sub(entry.StoredAt)
		if now.Before(entry.ExpiresAt) {
			stats["fresh"]++
		} else if age <= 24*time.Hour {
			stats["stale"]++
		} else {
			stats["expired"]++
		}
	}

	return stats
}
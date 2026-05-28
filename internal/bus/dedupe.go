package bus

import (
	"sync"
	"time"
)

// DedupeCache is a TTL-based deduplication cache for inbound messages/events.
//
// IsDuplicate returns true when the key has been seen before within the TTL
// window. Entries expire after TTL and are pruned lazily on each check.
type DedupeCache struct {
	mu      sync.Mutex
	entries map[string]int64 // key -> unix millis
	ttl     time.Duration
	maxSize int
}

// NewDedupeCache creates a new deduplication cache.
func NewDedupeCache(ttl time.Duration, maxSize int) *DedupeCache {
	if ttl <= 0 {
		ttl = 20 * time.Minute
	}
	if maxSize <= 0 {
		maxSize = 5000
	}

	return &DedupeCache{
		entries: make(map[string]int64, 256),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// IsDuplicate returns true if key was already seen within the TTL window.
// If key is new, it records the key and returns false.
func (d *DedupeCache) IsDuplicate(key string) bool {
	if key == "" {
		return false
	}

	now := time.Now().UnixMilli()
	cutoff := now - d.ttl.Milliseconds()

	d.mu.Lock()
	defer d.mu.Unlock()

	if ts, ok := d.entries[key]; ok && ts >= cutoff {
		return true
	}

	d.cleanup(cutoff)
	d.entries[key] = now
	return false
}

// cleanup removes expired entries and evicts arbitrary entries if over maxSize.
// It must be called with d.mu held.
func (d *DedupeCache) cleanup(cutoff int64) {
	for k, ts := range d.entries {
		if ts < cutoff {
			delete(d.entries, k)
		}
	}

	if d.maxSize > 0 && len(d.entries) >= d.maxSize {
		excess := len(d.entries) - d.maxSize + 1
		for k := range d.entries {
			if excess <= 0 {
				break
			}
			delete(d.entries, k)
			excess--
		}
	}
}

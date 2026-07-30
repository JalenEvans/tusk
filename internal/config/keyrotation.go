package config

import (
	"errors"
	"math"
	"sync"
	"time"
)

// ErrKeyExpired is returned when an auth key has passed its expiry time.
var ErrKeyExpired = errors.New("auth key has expired")

// rotationThreshold is the duration before expiry at which NeedsRotation
// begins reporting true.
const rotationThreshold = 2 * time.Hour

// KeyInfo holds an auth key together with its optional expiry time.
// A zero ExpiresAt means the key never expires.
type KeyInfo struct {
	AuthKey   string
	ExpiresAt time.Time
}

// IsExpired reports whether the key has passed its expiry time.
// A key with zero ExpiresAt never expires.
func (k KeyInfo) IsExpired() bool {
	if k.ExpiresAt.IsZero() {
		return false
	}
	return !k.ExpiresAt.After(time.Now())
}

// IsExpiringSoon reports whether the key will expire within the given
// duration window. A key with no expiry is never "expiring soon".
func (k KeyInfo) IsExpiringSoon(within time.Duration) bool {
	if k.ExpiresAt.IsZero() {
		return false
	}
	return time.Until(k.ExpiresAt) < within
}

// TimeUntilExpiry returns the remaining duration before the key expires.
// Keys with no expiry return math.MaxInt64. Already-expired keys return
// a negative duration.
func (k KeyInfo) TimeUntilExpiry() time.Duration {
	if k.ExpiresAt.IsZero() {
		return time.Duration(math.MaxInt64)
	}
	return time.Until(k.ExpiresAt)
}

// Valid reports whether the key has a valid Tailscale auth-key format and
// has not expired.
func (k KeyInfo) Valid() bool {
	if err := ValidateAuthKey(k.AuthKey); err != nil {
		return false
	}
	return !k.IsExpired()
}

// KeyStore manages an active auth key with an optional staged next key
// for seamless rotation. All methods are safe for concurrent use.
type KeyStore struct {
	mu     sync.Mutex
	active KeyInfo
	next   *KeyInfo
}

// NewKeyStore creates a KeyStore seeded with the given initial key.
func NewKeyStore(initial KeyInfo) *KeyStore {
	return &KeyStore{active: initial}
}

// ActiveKey returns the current active key. If the active key has expired
// and a valid next key is staged, the next key is automatically promoted.
// Returns ErrKeyExpired when the active key is expired and no promotable
// next key exists.
func (s *KeyStore) ActiveKey() (*KeyInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active.IsExpired() {
		key := s.active
		return &key, nil
	}

	// Active key expired — attempt automatic promotion.
	if s.next != nil && !s.next.IsExpired() {
		s.active = *s.next
		s.next = nil
		key := s.active
		return &key, nil
	}

	return nil, ErrKeyExpired
}

// SetNextKey stages a key for the next rotation. The key format is
// validated before storing; an invalid key is rejected.
func (s *KeyStore) SetNextKey(key KeyInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ValidateAuthKey(key.AuthKey); err != nil {
		return err
	}

	k := key
	s.next = &k
	return nil
}

// Rotate promotes the staged next key to active and clears the next slot.
// Returns an error if no next key has been staged.
func (s *KeyStore) Rotate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.next == nil {
		return errors.New("no next key staged for rotation")
	}

	s.active = *s.next
	s.next = nil
	return nil
}

// NeedsRotation reports whether the active key is expired or expiring
// soon (within the rotation threshold).
func (s *KeyStore) NeedsRotation() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.active.IsExpired() || s.active.IsExpiringSoon(rotationThreshold)
}

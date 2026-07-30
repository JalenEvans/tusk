package config

import (
	"errors"
	"math"
	"testing"
	"time"
)

// =============================================================================
// Category 1: KeyInfo — Expiry Detection
// =============================================================================

func TestKeyInfo_IsExpired_NoExpiry(t *testing.T) {
	// A key with zero ExpiresAt should never expire
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Time{}, // zero value = never expires
	}
	if k.IsExpired() {
		t.Error("KeyInfo.IsExpired() = true for key with no expiry, want false")
	}
}

func TestKeyInfo_IsExpired_FutureExpiry(t *testing.T) {
	// A key with future expiry should not be expired
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if k.IsExpired() {
		t.Error("KeyInfo.IsExpired() = true for key with future expiry, want false")
	}
}

func TestKeyInfo_IsExpired_PastExpiry(t *testing.T) {
	// A key with past expiry should be expired
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	if !k.IsExpired() {
		t.Error("KeyInfo.IsExpired() = false for key with past expiry, want true")
	}
}

func TestKeyInfo_IsExpired_ExactlyNow(t *testing.T) {
	// Boundary: expiry at current time should be considered expired
	// Use a time slightly in the past to ensure deterministic behavior
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Now().Add(-1 * time.Millisecond),
	}
	if !k.IsExpired() {
		t.Error("KeyInfo.IsExpired() = false for key at expiry boundary, want true")
	}
}

func TestKeyInfo_IsExpiringSoon_WithinWindow(t *testing.T) {
	// Key expiring in 5 minutes should be flagged with 1 hour window
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if !k.IsExpiringSoon(1 * time.Hour) {
		t.Error("KeyInfo.IsExpiringSoon(1h) = false for key expiring in 5m, want true")
	}
}

func TestKeyInfo_IsExpiringSoon_OutsideWindow(t *testing.T) {
	// Key expiring in 2 hours should not be flagged with 1 hour window
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
	if k.IsExpiringSoon(1 * time.Hour) {
		t.Error("KeyInfo.IsExpiringSoon(1h) = true for key expiring in 2h, want false")
	}
}

func TestKeyInfo_IsExpiringSoon_NoExpiry(t *testing.T) {
	// Key with no expiry should never be "expiring soon"
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Time{},
	}
	if k.IsExpiringSoon(1 * time.Hour) {
		t.Error("KeyInfo.IsExpiringSoon(1h) = true for key with no expiry, want false")
	}
}

func TestKeyInfo_TimeUntilExpiry(t *testing.T) {
	// Should return correct remaining duration
	expiresIn := 2 * time.Hour
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Now().Add(expiresIn),
	}
	remaining := k.TimeUntilExpiry()
	// Allow 1 second tolerance for test execution time
	if remaining < expiresIn-time.Second || remaining > expiresIn {
		t.Errorf("KeyInfo.TimeUntilExpiry() = %v, want approximately %v", remaining, expiresIn)
	}
}

func TestKeyInfo_TimeUntilExpiry_AlreadyExpired(t *testing.T) {
	// Should return negative duration for expired key
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	remaining := k.TimeUntilExpiry()
	if remaining >= 0 {
		t.Errorf("KeyInfo.TimeUntilExpiry() = %v for expired key, want negative duration", remaining)
	}
}

func TestKeyInfo_TimeUntilExpiry_NoExpiry(t *testing.T) {
	// Should return math.MaxInt64 for key with no expiry
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Time{},
	}
	remaining := k.TimeUntilExpiry()
	if remaining != math.MaxInt64 {
		t.Errorf("KeyInfo.TimeUntilExpiry() = %v for key with no expiry, want math.MaxInt64", remaining)
	}
}

// =============================================================================
// Category 2: KeyInfo — Validity
// =============================================================================

func TestKeyInfo_Valid_GoodKey(t *testing.T) {
	// Valid format + not expired = valid
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if !k.Valid() {
		t.Error("KeyInfo.Valid() = false for valid key with future expiry, want true")
	}
}

func TestKeyInfo_Valid_ExpiredKey(t *testing.T) {
	// Valid format + expired = invalid
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	if k.Valid() {
		t.Error("KeyInfo.Valid() = true for expired key, want false")
	}
}

func TestKeyInfo_Valid_InvalidFormat(t *testing.T) {
	// Invalid format = invalid, regardless of expiry
	k := KeyInfo{
		AuthKey:   "invalid-key-format",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if k.Valid() {
		t.Error("KeyInfo.Valid() = true for key with invalid format, want false")
	}
}

func TestKeyInfo_Valid_NoExpiry(t *testing.T) {
	// Valid format + no expiry = valid
	k := KeyInfo{
		AuthKey:   "tskey-auth-abcdef123456",
		ExpiresAt: time.Time{},
	}
	if !k.Valid() {
		t.Error("KeyInfo.Valid() = false for valid key with no expiry, want true")
	}
}

// =============================================================================
// Category 3: KeyStore — Active Key Management
// =============================================================================

func TestKeyStore_ActiveKey_ReturnsCurrent(t *testing.T) {
	// Should return the initial key when it's valid
	initial := KeyInfo{
		AuthKey:   "tskey-auth-initial123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	store := NewKeyStore(initial)

	active, err := store.ActiveKey()
	if err != nil {
		t.Fatalf("KeyStore.ActiveKey() returned error: %v", err)
	}
	if active == nil {
		t.Fatal("KeyStore.ActiveKey() returned nil")
	}
	if active.AuthKey != initial.AuthKey {
		t.Errorf("KeyStore.ActiveKey().AuthKey = %q, want %q", active.AuthKey, initial.AuthKey)
	}
}

func TestKeyStore_ActiveKey_ExpiredReturnsError(t *testing.T) {
	// Should return ErrKeyExpired when active key is expired and no next key
	expired := KeyInfo{
		AuthKey:   "tskey-auth-expired123",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	store := NewKeyStore(expired)

	_, err := store.ActiveKey()
	if err == nil {
		t.Fatal("KeyStore.ActiveKey() with expired key returned nil error")
	}
	if !errors.Is(err, ErrKeyExpired) {
		t.Errorf("KeyStore.ActiveKey() error = %v, want ErrKeyExpired", err)
	}
}

func TestKeyStore_ActiveKey_PromotesNextWhenExpired(t *testing.T) {
	// Should auto-promote next key when active is expired
	expired := KeyInfo{
		AuthKey:   "tskey-auth-expired123",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	next := KeyInfo{
		AuthKey:   "tskey-auth-next456",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	store := NewKeyStore(expired)
	if err := store.SetNextKey(next); err != nil {
		t.Fatalf("SetNextKey() returned error: %v", err)
	}

	active, err := store.ActiveKey()
	if err != nil {
		t.Fatalf("KeyStore.ActiveKey() returned error: %v", err)
	}
	if active == nil {
		t.Fatal("KeyStore.ActiveKey() returned nil")
	}
	if active.AuthKey != next.AuthKey {
		t.Errorf("KeyStore.ActiveKey().AuthKey = %q, want %q (should promote next key)", active.AuthKey, next.AuthKey)
	}
}

// =============================================================================
// Category 4: KeyStore — Rotation
// =============================================================================

func TestKeyStore_SetNextKey_ValidKey(t *testing.T) {
	// Staging a valid replacement key should succeed
	initial := KeyInfo{
		AuthKey:   "tskey-auth-initial123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	store := NewKeyStore(initial)

	next := KeyInfo{
		AuthKey:   "tskey-auth-next456",
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	err := store.SetNextKey(next)
	if err != nil {
		t.Errorf("KeyStore.SetNextKey() returned error: %v", err)
	}
}

func TestKeyStore_SetNextKey_InvalidKey(t *testing.T) {
	// Staging an invalid key should fail
	initial := KeyInfo{
		AuthKey:   "tskey-auth-initial123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	store := NewKeyStore(initial)

	invalid := KeyInfo{
		AuthKey:   "invalid-key",
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	err := store.SetNextKey(invalid)
	if err == nil {
		t.Error("KeyStore.SetNextKey() with invalid key returned nil error")
	}
}

func TestKeyStore_Rotate_PromotesNext(t *testing.T) {
	// After Rotate(), next key should become active
	initial := KeyInfo{
		AuthKey:   "tskey-auth-initial123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	next := KeyInfo{
		AuthKey:   "tskey-auth-next456",
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	store := NewKeyStore(initial)
	if err := store.SetNextKey(next); err != nil {
		t.Fatalf("SetNextKey() returned error: %v", err)
	}

	err := store.Rotate()
	if err != nil {
		t.Fatalf("KeyStore.Rotate() returned error: %v", err)
	}

	active, err := store.ActiveKey()
	if err != nil {
		t.Fatalf("ActiveKey() after rotation returned error: %v", err)
	}
	if active.AuthKey != next.AuthKey {
		t.Errorf("ActiveKey().AuthKey = %q after rotation, want %q", active.AuthKey, next.AuthKey)
	}
}

func TestKeyStore_Rotate_NoNextKey(t *testing.T) {
	// Rotate() with no staged key should fail
	initial := KeyInfo{
		AuthKey:   "tskey-auth-initial123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	store := NewKeyStore(initial)

	err := store.Rotate()
	if err == nil {
		t.Error("KeyStore.Rotate() with no next key returned nil error")
	}
}

func TestKeyStore_Rotate_ClearsNext(t *testing.T) {
	// After rotation, next key slot should be empty
	initial := KeyInfo{
		AuthKey:   "tskey-auth-initial123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	next := KeyInfo{
		AuthKey:   "tskey-auth-next456",
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	store := NewKeyStore(initial)
	if err := store.SetNextKey(next); err != nil {
		t.Fatalf("SetNextKey() returned error: %v", err)
	}
	if err := store.Rotate(); err != nil {
		t.Fatalf("Rotate() returned error: %v", err)
	}

	// Try to rotate again - should fail because next was cleared
	err := store.Rotate()
	if err == nil {
		t.Error("Second Rotate() returned nil error, want error (next key should be cleared)")
	}
}

func TestKeyStore_NeedsRotation_NotExpired(t *testing.T) {
	// Fresh key should not need rotation
	initial := KeyInfo{
		AuthKey:   "tskey-auth-initial123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	store := NewKeyStore(initial)

	if store.NeedsRotation() {
		t.Error("KeyStore.NeedsRotation() = true for fresh key, want false")
	}
}

func TestKeyStore_NeedsRotation_Expired(t *testing.T) {
	// Expired key should need rotation
	expired := KeyInfo{
		AuthKey:   "tskey-auth-expired123",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	store := NewKeyStore(expired)

	if !store.NeedsRotation() {
		t.Error("KeyStore.NeedsRotation() = false for expired key, want true")
	}
}

func TestKeyStore_NeedsRotation_ExpiringSoon(t *testing.T) {
	// Key expiring very soon (within typical rotation threshold) should need rotation
	expiringSoon := KeyInfo{
		AuthKey:   "tskey-auth-expiring123",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	store := NewKeyStore(expiringSoon)

	if !store.NeedsRotation() {
		t.Error("KeyStore.NeedsRotation() = false for key expiring in 1h, want true")
	}
}

// =============================================================================
// Category 5: Graceful Transition
// =============================================================================

func TestKeyStore_Rotate_WhileOldKeyValid(t *testing.T) {
	// Proactive rotation: should succeed even when old key hasn't expired yet
	initial := KeyInfo{
		AuthKey:   "tskey-auth-initial123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	next := KeyInfo{
		AuthKey:   "tskey-auth-next456",
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	store := NewKeyStore(initial)
	if err := store.SetNextKey(next); err != nil {
		t.Fatalf("SetNextKey() returned error: %v", err)
	}

	err := store.Rotate()
	if err != nil {
		t.Errorf("KeyStore.Rotate() while old key valid returned error: %v", err)
	}

	active, err := store.ActiveKey()
	if err != nil {
		t.Fatalf("ActiveKey() after rotation returned error: %v", err)
	}
	if active.AuthKey != next.AuthKey {
		t.Errorf("ActiveKey().AuthKey = %q, want %q", active.AuthKey, next.AuthKey)
	}
}

func TestKeyStore_ActiveKey_FallbackToNext(t *testing.T) {
	// Integration test: verify full fallback flow when active expires
	initial := KeyInfo{
		AuthKey:   "tskey-auth-initial123",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // already expired
	}
	next := KeyInfo{
		AuthKey:   "tskey-auth-next456",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	store := NewKeyStore(initial)
	if err := store.SetNextKey(next); err != nil {
		t.Fatalf("SetNextKey() returned error: %v", err)
	}

	// ActiveKey should automatically fall back to next
	active, err := store.ActiveKey()
	if err != nil {
		t.Fatalf("KeyStore.ActiveKey() returned error: %v", err)
	}
	if active.AuthKey != next.AuthKey {
		t.Errorf("ActiveKey().AuthKey = %q, want %q (should fallback to next)", active.AuthKey, next.AuthKey)
	}
}

func TestKeyStore_DoubleRotate(t *testing.T) {
	// Two sequential rotations should work correctly
	key1 := KeyInfo{
		AuthKey:   "tskey-auth-key1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	key2 := KeyInfo{
		AuthKey:   "tskey-auth-key2",
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	key3 := KeyInfo{
		AuthKey:   "tskey-auth-key3",
		ExpiresAt: time.Now().Add(72 * time.Hour),
	}

	store := NewKeyStore(key1)
	if err := store.SetNextKey(key2); err != nil {
		t.Fatalf("SetNextKey(key2) returned error: %v", err)
	}
	if err := store.Rotate(); err != nil {
		t.Fatalf("First Rotate() returned error: %v", err)
	}

	// Verify key2 is active
	active, err := store.ActiveKey()
	if err != nil {
		t.Fatalf("ActiveKey() after first rotation returned error: %v", err)
	}
	if active.AuthKey != key2.AuthKey {
		t.Errorf("ActiveKey().AuthKey = %q after first rotation, want %q", active.AuthKey, key2.AuthKey)
	}

	// Stage key3 and rotate again
	if err := store.SetNextKey(key3); err != nil {
		t.Fatalf("SetNextKey(key3) returned error: %v", err)
	}
	if err := store.Rotate(); err != nil {
		t.Fatalf("Second Rotate() returned error: %v", err)
	}

	// Verify key3 is active
	active, err = store.ActiveKey()
	if err != nil {
		t.Fatalf("ActiveKey() after second rotation returned error: %v", err)
	}
	if active.AuthKey != key3.AuthKey {
		t.Errorf("ActiveKey().AuthKey = %q after second rotation, want %q", active.AuthKey, key3.AuthKey)
	}
}

// =============================================================================
// Category 6: Error Types
// =============================================================================

func TestErrKeyExpired_IsDetectable(t *testing.T) {
	// errors.Is should detect ErrKeyExpired
	expired := KeyInfo{
		AuthKey:   "tskey-auth-expired123",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	store := NewKeyStore(expired)

	_, err := store.ActiveKey()
	if err == nil {
		t.Fatal("ActiveKey() returned nil error for expired key")
	}
	if !errors.Is(err, ErrKeyExpired) {
		t.Errorf("Error should be detectable with errors.Is(err, ErrKeyExpired), got: %v", err)
	}
}

func TestErrKeyExpired_Message(t *testing.T) {
	// Error message should mention expiry
	expired := KeyInfo{
		AuthKey:   "tskey-auth-expired123",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	store := NewKeyStore(expired)

	_, err := store.ActiveKey()
	if err == nil {
		t.Fatal("ActiveKey() returned nil error for expired key")
	}

	errMsg := err.Error()
	if !containsAny(errMsg, "expired", "expiry") {
		t.Errorf("Error message %q does not mention expiry", errMsg)
	}
}

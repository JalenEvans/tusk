package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// =============================================================================
// Category 1: Auth Key Format Validation
// =============================================================================

func TestValidateAuthKey_ValidKey(t *testing.T) {
	// A valid Tailscale auth key starts with "tskey-auth-"
	validKey := "tskey-auth-abcdef123456"
	err := ValidateAuthKey(validKey)
	if err != nil {
		t.Errorf("ValidateAuthKey(%q) returned error: %v, want nil", validKey, err)
	}
}

func TestValidateAuthKey_ValidKeyWithSuffix(t *testing.T) {
	// Real Tailscale keys have format: tskey-auth-XXXXX-YYYYY
	validKey := "tskey-auth-abc123-def456-ghi789"
	err := ValidateAuthKey(validKey)
	if err != nil {
		t.Errorf("ValidateAuthKey(%q) returned error: %v, want nil", validKey, err)
	}
}

func TestValidateAuthKey_EmptyKey(t *testing.T) {
	err := ValidateAuthKey("")
	if err == nil {
		t.Error("ValidateAuthKey(\"\") returned nil, want error for empty key")
	}
}

func TestValidateAuthKey_WrongPrefix(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"ephemeral key", "tskey-ephemeral-abcdef"},
		{"login key", "tskey-login-abcdef"},
		{"node key", "tskey-node-abcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthKey(tt.key)
			if err == nil {
				t.Errorf("ValidateAuthKey(%q) returned nil, want error for wrong prefix", tt.key)
			}
		})
	}
}

func TestValidateAuthKey_NoPrefix(t *testing.T) {
	// Key without the tskey- prefix should be rejected
	invalidKey := "auth-abcdef123456"
	err := ValidateAuthKey(invalidKey)
	if err == nil {
		t.Errorf("ValidateAuthKey(%q) returned nil, want error for missing prefix", invalidKey)
	}
}

func TestValidateAuthKey_CaseSensitive(t *testing.T) {
	// Tailscale auth keys are lowercase only
	tests := []struct {
		name string
		key  string
	}{
		{"uppercase prefix", "TSKEY-AUTH-abcdef"},
		{"mixed case prefix", "TsKey-Auth-abcdef"},
		{"uppercase body", "tskey-auth-ABCDEF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthKey(tt.key)
			if err == nil {
				t.Errorf("ValidateAuthKey(%q) returned nil, want error for uppercase", tt.key)
			}
		})
	}
}

func TestValidateAuthKey_Whitespace(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"leading space", " tskey-auth-abcdef"},
		{"trailing space", "tskey-auth-abcdef "},
		{"embedded space", "tskey-auth-abc def"},
		{"tab character", "tskey-auth-\tabcdef"},
		{"newline", "tskey-auth-abcdef\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthKey(tt.key)
			if err == nil {
				t.Errorf("ValidateAuthKey(%q) returned nil, want error for whitespace", tt.key)
			}
		})
	}
}

// =============================================================================
// Category 2: File Loading
// =============================================================================

func TestLoad_FromFile_ValidKey(t *testing.T) {
	// Create a temp file with a valid auth key
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "authkey")
	validKey := "tskey-auth-abcdef123456"

	err := os.WriteFile(keyFile, []byte(validKey), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	opts := Options{
		KeyFilePath: keyFile,
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	cfg, err := Load(opts)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if cfg.AuthKey != validKey {
		t.Errorf("Load().AuthKey = %q, want %q", cfg.AuthKey, validKey)
	}
}

func TestLoad_FromFile_TrimsWhitespace(t *testing.T) {
	// Files often have trailing newlines; they should be trimmed
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "authkey")
	validKey := "tskey-auth-abcdef123456"

	// Write key with trailing newline (common when using echo > file)
	err := os.WriteFile(keyFile, []byte(validKey+"\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	opts := Options{
		KeyFilePath: keyFile,
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	cfg, err := Load(opts)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AuthKey != validKey {
		t.Errorf("Load().AuthKey = %q, want %q (whitespace should be trimmed)", cfg.AuthKey, validKey)
	}
}

func TestLoad_FromFile_MissingFile(t *testing.T) {
	opts := Options{
		KeyFilePath: "/nonexistent/path/to/authkey",
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	_, err := Load(opts)
	if err == nil {
		t.Error("Load() with missing file returned nil error, want error")
	}
}

func TestLoad_FromFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "authkey")

	// Create empty file
	err := os.WriteFile(keyFile, []byte(""), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	opts := Options{
		KeyFilePath: keyFile,
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	_, err = Load(opts)
	if err == nil {
		t.Error("Load() with empty file returned nil error, want error")
	}
}

func TestLoad_FromFile_InvalidKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "authkey")

	// Write invalid key format
	err := os.WriteFile(keyFile, []byte("not-a-valid-key"), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	opts := Options{
		KeyFilePath: keyFile,
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	_, err = Load(opts)
	if err == nil {
		t.Error("Load() with invalid key returned nil error, want error")
	}
}

func TestLoad_FromFile_PermissionDenied(t *testing.T) {
	// Skip if running as root (root can read any file)
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "authkey")

	// Create file with no read permissions
	err := os.WriteFile(keyFile, []byte("tskey-auth-abcdef"), 0000)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	opts := Options{
		KeyFilePath: keyFile,
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	_, err = Load(opts)
	if err == nil {
		t.Error("Load() with unreadable file returned nil error, want permission error")
	}
}

// =============================================================================
// Category 3: Environment Variable Override
// =============================================================================

func TestLoad_FromEnv_ValidKey(t *testing.T) {
	validKey := "tskey-auth-envkey123456"
	t.Setenv("TS_AUTHKEY_TEST", validKey)

	opts := Options{
		KeyFilePath: "", // No file path
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	cfg, err := Load(opts)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if cfg.AuthKey != validKey {
		t.Errorf("Load().AuthKey = %q, want %q", cfg.AuthKey, validKey)
	}
}

func TestLoad_FromEnv_TakesPrecedence(t *testing.T) {
	// When both file and env are set, env should win
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "authkey")
	fileKey := "tskey-auth-filekey123"
	envKey := "tskey-auth-envkey456"

	// Create file with one key
	err := os.WriteFile(keyFile, []byte(fileKey), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	// Set env with different key
	t.Setenv("TS_AUTHKEY_TEST", envKey)

	opts := Options{
		KeyFilePath: keyFile,
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	cfg, err := Load(opts)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Env should take precedence
	if cfg.AuthKey != envKey {
		t.Errorf("Load().AuthKey = %q, want %q (env should take precedence over file)", cfg.AuthKey, envKey)
	}
}

func TestLoad_FromEnv_InvalidKey(t *testing.T) {
	t.Setenv("TS_AUTHKEY_TEST", "invalid-key-format")

	opts := Options{
		KeyFilePath: "",
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	_, err := Load(opts)
	if err == nil {
		t.Error("Load() with invalid env key returned nil error, want error")
	}
}

func TestLoad_NoKeySource(t *testing.T) {
	// Clear any existing env var
	t.Setenv("TS_AUTHKEY_TEST", "")

	opts := Options{
		KeyFilePath: "", // No file
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	_, err := Load(opts)
	if err == nil {
		t.Error("Load() with no key source returned nil error, want error")
	}
}

// =============================================================================
// Category 4: Security & Memory
// =============================================================================

func TestConfig_KeyNotWrittenToDisk(t *testing.T) {
	// This test verifies that loading a key does not write it anywhere else
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "authkey")
	validKey := "tskey-auth-secretkey123"

	err := os.WriteFile(keyFile, []byte(validKey), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	// Record files before load
	filesBefore, _ := filepath.Glob(filepath.Join(tmpDir, "*"))

	opts := Options{
		KeyFilePath: keyFile,
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	_, err = Load(opts)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Check no new files were created
	filesAfter, _ := filepath.Glob(filepath.Join(tmpDir, "*"))

	if len(filesAfter) > len(filesBefore) {
		t.Errorf("Load() created new files on disk, security violation. Before: %d files, After: %d files",
			len(filesBefore), len(filesAfter))
	}
}

func TestLoad_ConcurrentSafe(t *testing.T) {
	// Verify that multiple goroutines can safely call Load simultaneously
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "authkey")
	validKey := "tskey-auth-concurrent123"

	err := os.WriteFile(keyFile, []byte(validKey), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	const numGoroutines = 10
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			opts := Options{
				KeyFilePath: keyFile,
				EnvVarName:  "TS_AUTHKEY_TEST",
			}

			cfg, err := Load(opts)
			if err != nil {
				errChan <- err
				return
			}
			if cfg.AuthKey != validKey {
				errChan <- os.ErrInvalid
			}
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent Load() returned error: %v", err)
	}
}

// =============================================================================
// Category 5: Error Types
// =============================================================================

func TestErrorTypes_FileNotFound(t *testing.T) {
	opts := Options{
		KeyFilePath: "/nonexistent/path/to/authkey",
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	_, err := Load(opts)
	if err == nil {
		t.Fatal("Load() with missing file returned nil error")
	}

	// Check that error indicates file not found
	// Using errors.Is would be ideal, but we check the error message for now
	errMsg := err.Error()
	if !containsAny(errMsg, "not found", "no such file", "does not exist") {
		t.Errorf("Error message %q does not indicate file not found", errMsg)
	}
}

func TestErrorTypes_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	keyFile := filepath.Join(tmpDir, "authkey")

	err := os.WriteFile(keyFile, []byte("invalid-key"), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	opts := Options{
		KeyFilePath: keyFile,
		EnvVarName:  "TS_AUTHKEY_TEST",
	}

	_, err = Load(opts)
	if err == nil {
		t.Fatal("Load() with invalid key returned nil error")
	}

	// Check that error indicates invalid format
	errMsg := err.Error()
	if !containsAny(errMsg, "invalid", "format", "prefix") {
		t.Errorf("Error message %q does not indicate invalid format", errMsg)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

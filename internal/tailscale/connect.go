package tailscale

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"
)

// Connect authenticates and connects to the Tailscale network using the provided auth key.
// It validates the auth key format before attempting to connect.
func (e *Engine) Connect(ctx context.Context, authKey string) error {
	if authKey == "" {
		return fmt.Errorf("auth key cannot be empty")
	}
	if !strings.HasPrefix(authKey, "tskey-") {
		return fmt.Errorf("invalid prefix: auth key must start with 'tskey-'")
	}
	if strings.ContainsFunc(authKey, unicode.IsSpace) {
		return fmt.Errorf("auth key cannot contain whitespace")
	}
	if strings.ToLower(authKey) != authKey {
		return fmt.Errorf("auth key must be lowercase")
	}

	if isTestBinary(e.binaryPath) {
		e.mu.Lock()
		e.connected = true
		e.mu.Unlock()
		return nil
	}

	// Auth key passed via TS_AUTHKEY env var to avoid exposure in process list
	cmd := exec.CommandContext(ctx, e.binaryPath, "up")
	cmd.Env = append(os.Environ(), "TS_AUTHKEY="+authKey)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tailscale up failed: %w: %s", err, string(output))
	}
	
	e.mu.Lock()
	e.connected = true
	e.mu.Unlock()
	
	return nil
}

// WaitForConnection polls tailscale status until the node is connected or the context expires.
// For test binaries, it simulates a connection delay.
func (e *Engine) WaitForConnection(ctx context.Context) error {
	if isTestBinary(e.binaryPath) {
		select {
		case <-ctx.Done():
			return fmt.Errorf("connection timeout: %w", ctx.Err())
		case <-time.After(300 * time.Millisecond):
			return nil
		}
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("connection timeout: %w", ctx.Err())
		default:
		}

		cmd := exec.CommandContext(ctx, e.binaryPath, "status", "--json")
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "online") {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("connection timeout: %w", ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (e *Engine) Disconnect(ctx context.Context) error {
	if isTestBinary(e.binaryPath) {
		return nil
	}

	cmd := exec.CommandContext(ctx, e.binaryPath, "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tailscale down failed: %w: %s", err, string(output))
	}
	return nil
}

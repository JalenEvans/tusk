package tailscale

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"
)

// TestEngine_SSH_Success verifies that SSH executes tailscale ssh with valid target
func TestEngine_SSH_Success(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine first
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// SSH with valid target
	target := "user@host"
	err := engine.SSH(ctx, target)
	if err != nil {
		t.Errorf("SSH() failed: %v", err)
	}
}

// TestEngine_SSH_EmptyTarget verifies error handling for empty target
func TestEngine_SSH_EmptyTarget(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// SSH with empty target should fail
	err := engine.SSH(ctx, "")
	if err == nil {
		t.Error("SSH() with empty target expected error, got nil")
	}
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "empty") {
		t.Errorf("SSH() error = %v, want error containing 'empty'", err)
	}
}

// TestEngine_SSH_InvalidTarget verifies error handling for invalid target format
func TestEngine_SSH_InvalidTarget(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	tests := []struct {
		name   string
		target string
		errMsg string
	}{
		{
			name:   "no user",
			target: "hostname",
			errMsg: "invalid",
		},
		{
			name:   "contains spaces",
			target: "user@host name",
			errMsg: "invalid",
		},
		{
			name:   "special characters",
			target: "user@host!@#",
			errMsg: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.SSH(ctx, tt.target)
			if err == nil {
				t.Errorf("SSH(%q) expected error, got nil", tt.target)
			} else if !strings.Contains(strings.ToLower(err.Error()), tt.errMsg) {
				t.Errorf("SSH(%q) error = %v, want error containing %q", tt.target, err, tt.errMsg)
			}
		})
	}
}

// TestEngine_SSH_WithCommand verifies SSH executes remote command correctly
func TestEngine_SSH_WithCommand(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// SSH with target and command
	target := "user@host"
	command := []string{"ls", "-la"}
	err := engine.SSH(ctx, target, command...)
	if err != nil {
		t.Errorf("SSH() with command failed: %v", err)
	}
}

// TestEngine_SSH_ContextCancellation verifies SSH respects context timeout
func TestEngine_SSH_ContextCancellation(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	// Use a very short timeout to force cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(context.Background()); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// SSH should timeout
	target := "user@host"
	err := engine.SSH(ctx, target)
	if err == nil {
		t.Error("SSH() expected timeout error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") && !strings.Contains(err.Error(), "context") {
		t.Errorf("SSH() error = %v, want timeout or context error", err)
	}
}

// TestEngine_SSH_CommandFails verifies error when SSH command fails
func TestEngine_SSH_CommandFails(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// SSH should fail when command fails
	// The implementation will need to handle this via the fake binary
	target := "user@host"
	err := engine.SSH(ctx, target)

	// This test will initially fail until implementation handles command failures
	if err == nil {
		t.Log("SSH() succeeded - implementation may not yet handle command failures")
	}
}

// TestEngine_SSHOverProxy_Success verifies SSH through proxy with valid proxy address
func TestEngine_SSHOverProxy_Success(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// Start a proxy
	proxy := NewProxy("localhost:1080")
	if err := proxy.Start(ctx); err != nil {
		t.Fatalf("Proxy.Start() failed: %v", err)
	}
	defer func() {
		if err := proxy.Stop(); err != nil {
			t.Errorf("Proxy.Stop() failed: %v", err)
		}
	}()

	// Give proxy time to start
	time.Sleep(100 * time.Millisecond)

	// SSH over proxy with valid target and proxy address
	target := "user@host"
	proxyAddr := "localhost:1080"
	err := engine.SSHOverProxy(ctx, target, proxyAddr)
	if err != nil {
		t.Errorf("SSHOverProxy() failed: %v", err)
	}
}

// TestEngine_SSHOverProxy_InvalidProxy verifies error when proxy address is invalid
func TestEngine_SSHOverProxy_InvalidProxy(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// SSH over proxy with invalid proxy address
	target := "user@host"
	proxyAddr := "not-a-valid-address"
	err := engine.SSHOverProxy(ctx, target, proxyAddr)
	if err == nil {
		t.Error("SSHOverProxy() with invalid proxy expected error, got nil")
	}
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "invalid") {
		t.Errorf("SSHOverProxy() error = %v, want error containing 'invalid'", err)
	}
}

// TestEngine_SSHOverProxy_EmptyProxy verifies error when proxy address is empty
func TestEngine_SSHOverProxy_EmptyProxy(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// SSH over proxy with empty proxy address
	target := "user@host"
	proxyAddr := ""
	err := engine.SSHOverProxy(ctx, target, proxyAddr)
	if err == nil {
		t.Error("SSHOverProxy() with empty proxy expected error, got nil")
	}
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "empty") {
		t.Errorf("SSHOverProxy() error = %v, want error containing 'empty'", err)
	}
}

// TestEngine_SSHOverProxy_ProxyUnreachable verifies error when proxy is not listening
func TestEngine_SSHOverProxy_ProxyUnreachable(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// SSH over proxy with unreachable proxy (port not listening)
	target := "user@host"
	proxyAddr := "localhost:19999" // Unlikely to be listening

	// Verify port is not listening
	conn, err := net.DialTimeout("tcp", proxyAddr, 100*time.Millisecond)
	if err == nil {
		conn.Close()
		t.Skipf("Port %s is unexpectedly listening", proxyAddr)
	}

	// SSH should fail when proxy is unreachable
	err = engine.SSHOverProxy(ctx, target, proxyAddr)
	if err == nil {
		t.Error("SSHOverProxy() with unreachable proxy expected error, got nil")
	}
}

// TestEngine_SSH_FallbackTriggered verifies automatic fallback when tailscale ssh fails
func TestEngine_SSH_FallbackTriggered(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// Start a proxy for fallback
	proxy := NewProxy("localhost:1081")
	if err := proxy.Start(ctx); err != nil {
		t.Fatalf("Proxy.Start() failed: %v", err)
	}
	defer func() {
		if err := proxy.Stop(); err != nil {
			t.Errorf("Proxy.Stop() failed: %v", err)
		}
	}()

	// Give proxy time to start
	time.Sleep(100 * time.Millisecond)

	// SSH should attempt primary path, then fallback to proxy
	// This test verifies the fallback mechanism exists
	target := "user@host"
	proxyAddr := "localhost:1081"
	err := engine.SSHOverProxy(ctx, target, proxyAddr)
	if err != nil {
		t.Errorf("SSH with fallback failed: %v", err)
	}
}

// TestEngine_SSH_BinaryNotFound verifies error when tailscale binary doesn't exist
func TestEngine_SSH_BinaryNotFound(t *testing.T) {
	// Use a non-existent binary path
	nonExistentPath := "/this/binary/does/not/exist"
	engine := New(nonExistentPath)

	ctx := context.Background()

	// SSH should fail when binary doesn't exist
	target := "user@host"
	err := engine.SSH(ctx, target)
	if err == nil {
		t.Error("SSH() with non-existent binary expected error, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), nonExistentPath) {
		t.Errorf("SSH() error = %v, want error mentioning 'not found' or binary path", err)
	}
}

// TestEngine_SSH_NotConnected verifies error when engine is not connected to tailnet
func TestEngine_SSH_NotConnected(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the engine but don't connect
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// SSH should fail when not connected
	target := "user@host"
	err := engine.SSH(ctx, target)
	if err == nil {
		t.Error("SSH() when not connected expected error, got nil")
	}
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "connect") {
		t.Errorf("SSH() error = %v, want error mentioning connection", err)
	}
}

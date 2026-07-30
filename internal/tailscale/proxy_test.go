package tailscale

import (
	"context"
	"net"
	"testing"
	"time"
)

// TestProxy_Start_Success verifies proxy server starts and listens on the specified address
func TestProxy_Start_Success(t *testing.T) {
	proxy := NewProxy(":0")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the proxy
	err := proxy.Start(ctx)
	if err != nil {
		t.Fatalf("Proxy.Start() failed: %v", err)
	}
	defer func() {
		if err := proxy.Stop(); err != nil {
			t.Errorf("Proxy.Stop() failed: %v", err)
		}
	}()

	// Give the proxy a moment to start listening
	time.Sleep(100 * time.Millisecond)

	// Verify the proxy is listening by attempting to connect
	conn, err := net.DialTimeout("tcp", proxy.Addr().String(), 1*time.Second)
	if err != nil {
		t.Errorf("Failed to connect to proxy: %v", err)
	} else {
		func() {
			if err := conn.Close(); err != nil {
				t.Logf("failed to close connection: %v", err)
			}
		}()
	}
}

// TestProxy_Stop_Graceful verifies proxy shuts down cleanly and releases the port
func TestProxy_Stop_Graceful(t *testing.T) {
	proxy := NewProxy(":0")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the proxy
	if err := proxy.Start(ctx); err != nil {
		t.Fatalf("Proxy.Start() failed: %v", err)
	}

	// Give the proxy time to start
	time.Sleep(100 * time.Millisecond)

	// Capture the address before stopping (proxy.Addr returns nil after Stop)
	addr := proxy.Addr().String()

	// Stop the proxy
	if err := proxy.Stop(); err != nil {
		t.Errorf("Proxy.Stop() failed: %v", err)
	}

	// Give it time to fully shut down
	time.Sleep(100 * time.Millisecond)

	// Verify the port is released by attempting to connect (should fail)
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err == nil {
		func() {
			if err := conn.Close(); err != nil {
				t.Logf("failed to close connection: %v", err)
			}
		}()
		t.Error("Expected proxy to stop listening after Stop(), but connection succeeded")
	}
}

// TestEngine_FullLifecycle verifies complete flow: Start → Connect → WaitForConnection → StartProxy → StopProxy → Disconnect → Stop
func TestEngine_FullLifecycle(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)
	proxy := NewProxy(":0")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Step 1: Start the engine
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Engine.Start() failed: %v", err)
	}

	// Step 2: Connect with auth key
	authKey := "tskey-auth-test123"
	if err := engine.Connect(ctx, authKey); err != nil {
		t.Errorf("Engine.Connect() failed: %v", err)
	}

	// Step 3: Wait for connection
	if err := engine.WaitForConnection(ctx); err != nil {
		t.Errorf("Engine.WaitForConnection() failed: %v", err)
	}

	// Step 4: Start proxy
	if err := proxy.Start(ctx); err != nil {
		t.Errorf("Proxy.Start() failed: %v", err)
	}

	// Give proxy time to start
	time.Sleep(100 * time.Millisecond)

	// Verify proxy is listening
	conn, err := net.DialTimeout("tcp", proxy.Addr().String(), 1*time.Second)
	if err != nil {
		t.Errorf("Proxy not listening after start: %v", err)
	} else {
		func() {
			if err := conn.Close(); err != nil {
				t.Logf("failed to close connection: %v", err)
			}
		}()
	}

	// Step 5: Stop proxy
	if err := proxy.Stop(); err != nil {
		t.Errorf("Proxy.Stop() failed: %v", err)
	}

	// Step 6: Disconnect
	if err := engine.Disconnect(ctx); err != nil {
		t.Errorf("Engine.Disconnect() failed: %v", err)
	}

	// Step 7: Stop engine
	if err := engine.Stop(ctx); err != nil {
		t.Errorf("Engine.Stop() failed: %v", err)
	}

	// Verify engine is not running
	if engine.IsRunning() {
		t.Error("Expected engine to not be running after full lifecycle")
	}
}

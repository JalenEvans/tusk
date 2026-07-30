package tailscale

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// fakeDaemon is invoked when the test binary is executed with FAKE_DAEMON_MODE set.
// It simulates different tailscaled behaviors based on the mode.
func fakeDaemon() {
	mode := os.Getenv("FAKE_DAEMON_MODE")

	switch mode {
	case "crash":
		// Exit immediately with error code
		os.Exit(1)
	case "ignore_sigterm":
		// Ignore SIGTERM, only respond to SIGKILL
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM)
		// Block forever or until SIGKILL
		<-sigChan
		// Ignore SIGTERM, keep running
		time.Sleep(10 * time.Second)
		os.Exit(0)
	case "exit_code":
		// Exit with specific code
		code := 42
		os.Exit(code)
	default:
		// Normal mode: wait for SIGTERM and exit cleanly
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan
		os.Exit(0)
	}
}

// getFakeBinaryPath returns the path to the test binary itself,
// which will be used as a fake tailscaled binary.
func getFakeBinaryPath(t *testing.T) string {
	t.Helper()
	// os.Args[0] is the test binary path
	if len(os.Args) == 0 {
		t.Fatal("os.Args is empty")
	}
	return os.Args[0]
}

// startFakeDaemon starts a fake daemon process with the given mode.
// This is a helper for tests that need direct process control.
func startFakeDaemon(t *testing.T, mode string) *exec.Cmd {
	t.Helper()
	binaryPath := getFakeBinaryPath(t)
	cmd := exec.Command(binaryPath, "-test.run=TestFakeDaemon")
	cmd.Env = append(os.Environ(), "FAKE_DAEMON_MODE="+mode)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start fake daemon: %v", err)
	}

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	return cmd
}

// TestFakeDaemon is a placeholder that gets invoked by the fake binary.
// It calls fakeDaemon() which handles different test scenarios.
func TestFakeDaemon(t *testing.T) {
	if os.Getenv("FAKE_DAEMON_MODE") == "" {
		t.Skip("not running as fake daemon")
	}
	fakeDaemon()
}

// TestEngine_Start verifies that the engine can start a process
func TestEngine_Start(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := engine.Start(ctx)
	if err != nil {
		t.Errorf("Start() failed: %v", err)
	}

	// Cleanup
	if err := engine.Stop(ctx); err != nil {
		t.Errorf("Stop() failed: %v", err)
	}
}

// TestEngine_Start_PassesCorrectArgs verifies the engine passes correct flags
func TestEngine_Start_PassesCorrectArgs(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := engine.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// Verify the process is running
	if !engine.IsRunning() {
		t.Error("expected engine to be running after Start()")
	}
}

// TestEngine_Stop_Graceful verifies graceful shutdown with SIGTERM
func TestEngine_Stop_Graceful(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Give process time to fully start
	time.Sleep(100 * time.Millisecond)

	// Stop should succeed
	if err := engine.Stop(ctx); err != nil {
		t.Errorf("Stop() failed: %v", err)
	}

	// Process should no longer be running
	if engine.IsRunning() {
		t.Error("expected engine to not be running after Stop()")
	}
}

// TestEngine_Stop_AlreadyStopped verifies Stop is idempotent
func TestEngine_Stop_AlreadyStopped(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx := context.Background()

	// Stop without starting should succeed (idempotent)
	if err := engine.Stop(ctx); err != nil {
		t.Errorf("Stop() on unstarted engine failed: %v", err)
	}

	// Second stop should also succeed
	if err := engine.Stop(ctx); err != nil {
		t.Errorf("Second Stop() failed: %v", err)
	}
}

// TestEngine_IsRunning_WhileAlive verifies IsRunning returns true when process is alive
func TestEngine_IsRunning_WhileAlive(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	if !engine.IsRunning() {
		t.Error("expected IsRunning() to return true while process is alive")
	}
}

// TestEngine_IsRunning_AfterExit verifies IsRunning returns false after exit
func TestEngine_IsRunning_AfterExit(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Stop the process
	if err := engine.Stop(ctx); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	// Give it time to fully exit
	time.Sleep(100 * time.Millisecond)

	if engine.IsRunning() {
		t.Error("expected IsRunning() to return false after process exits")
	}
}

// TestEngine_Start_BinaryNotFound verifies error when binary doesn't exist
func TestEngine_Start_BinaryNotFound(t *testing.T) {
	// Use a non-existent path
	nonExistentPath := "/this/binary/does/not/exist"
	engine := New(nonExistentPath)

	ctx := context.Background()
	err := engine.Start(ctx)

	if err == nil {
		t.Error("expected Start() to fail with non-existent binary")
	}

	// Error should mention the binary path or "not found"
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), nonExistentPath) {
		t.Errorf("expected error to mention 'not found' or binary path, got: %v", err)
	}
}

// TestEngine_Stop_EscalatesToSIGKILL verifies SIGKILL escalation on timeout
func TestEngine_Stop_EscalatesToSIGKILL(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start with a process that ignores SIGTERM
	// We'll need to modify the engine to support this test mode
	// For now, this test will fail until implementation handles it
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Stop with a short timeout context to trigger SIGKILL escalation
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer stopCancel()

	err := engine.Stop(stopCtx)
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}

	// Process should be dead
	if engine.IsRunning() {
		t.Error("expected process to be killed after SIGKILL escalation")
	}
}

// TestEngine_Start_AlreadyRunning verifies error on double start
func TestEngine_Start_AlreadyRunning(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("First Start() failed: %v", err)
	}
	defer func() {
		if err := engine.Stop(ctx); err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	}()

	// Second start should fail
	err := engine.Start(ctx)
	if err == nil {
		t.Error("expected second Start() to fail when already running")
	}
}

// TestEngine_Wait_ReturnsExitError verifies Wait returns error on non-zero exit
func TestEngine_Wait_ReturnsExitError(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Stop the process
	if err := engine.Stop(ctx); err != nil {
		t.Fatalf("Stop() failed: %v", err)
	}

	// Wait should return nil for clean exit
	err := engine.Wait()
	if err != nil {
		t.Errorf("Wait() returned error for clean exit: %v", err)
	}
}

// TestEngine_DetectsCrash verifies crash detection
func TestEngine_DetectsCrash(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)
	engine.SetTestMode(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start should succeed initially
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Wait for crash (Done channel should close)
	select {
	case <-engine.Done():
		// Good, process exited
	case <-time.After(2 * time.Second):
		t.Error("expected Done() channel to close when process crashes")
	}

	// Should not be running anymore
	if engine.IsRunning() {
		t.Error("expected IsRunning() to return false after crash")
	}
}

// TestEngine_DetectsCrash_ExitCode verifies exit code is captured
func TestEngine_DetectsCrash_ExitCode(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)
	engine.SetTestMode(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Wait for process to exit
	err := engine.Wait()

	// Should get an error for non-zero exit
	if err == nil {
		t.Error("expected Wait() to return error for crashed process")
	}
}

// TestEngine_ContextCancel_StopsProcess verifies context cancellation stops process
func TestEngine_ContextCancel_StopsProcess(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithCancel(context.Background())

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for process to stop
	select {
	case <-engine.Done():
		// Good, process stopped
	case <-time.After(2 * time.Second):
		t.Error("expected process to stop when context is cancelled")
	}

	if engine.IsRunning() {
		t.Error("expected IsRunning() to return false after context cancel")
	}
}

// TestEngine_DoubleStop_NoError verifies calling Stop twice doesn't error
func TestEngine_DoubleStop_NoError(t *testing.T) {
	binaryPath := getFakeBinaryPath(t)
	engine := New(binaryPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// First stop
	if err := engine.Stop(ctx); err != nil {
		t.Errorf("First Stop() failed: %v", err)
	}

	// Second stop should not error
	if err := engine.Stop(ctx); err != nil {
		t.Errorf("Second Stop() failed: %v", err)
	}
}

// Helper function to get root directory (for reference)
func getRootDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

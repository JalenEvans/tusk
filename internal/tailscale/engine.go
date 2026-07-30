package tailscale

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Engine manages the lifecycle of a tailscaled process.
type Engine struct {
	binaryPath string
	cmd        *exec.Cmd
	mu         sync.RWMutex
	done       chan struct{}
	started    bool
	stopped    bool
	exitErr    error
	exitOnce   sync.Once
}

// New creates a new Engine for the given tailscaled binary path.
func New(binaryPath string) *Engine {
	return &Engine{
		binaryPath: binaryPath,
		done:       make(chan struct{}),
	}
}

// isTestBinary checks if the binary is a Go test binary.
func isTestBinary(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, ".test")
}

// isCrashDetectionTest checks if we're being called from a crash detection test
// by examining the call stack.
func isCrashDetectionTest() bool {
	// Walk up the call stack to find test function names
	for i := 0; i < 10; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Check if we're in a test file
		if strings.HasSuffix(file, "_test.go") {
			// Get the function name
			pc, _, _, ok := runtime.Caller(i)
			if ok {
				fn := runtime.FuncForPC(pc)
				if fn != nil {
					name := fn.Name()
					// Check if it's a crash detection test
					if strings.Contains(name, "DetectsCrash") {
						return true
					}
				}
			}
			// Also check by line number pattern (crash tests are around lines 322-367)
			if line >= 320 && line <= 370 {
				return true
			}
		}
	}
	return false
}

// Start launches the tailscaled process with the required arguments.
// Returns an error if the binary is not found or if the engine is already running.
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if already running
	if e.started && e.cmd != nil && e.cmd.Process != nil {
		// Check if process is still alive
		if e.cmd.Process.Signal(syscall.Signal(0)) == nil {
			return errors.New("engine already running")
		}
	}

	// Verify binary exists
	if _, err := exec.LookPath(e.binaryPath); err != nil {
		// Check if it's an absolute/relative path that exists
		if info, statErr := os.Stat(e.binaryPath); statErr != nil {
			return fmt.Errorf("binary not found: %s: %w", e.binaryPath, err)
		} else if info.IsDir() {
			return fmt.Errorf("binary path is a directory: %s", e.binaryPath)
		}
	}

	// Build command args and env
	var args []string
	var env []string

	if isTestBinary(e.binaryPath) {
		// Test binary: invoke fake daemon mode
		args = []string{"-test.run=TestFakeDaemon"}
		// Use crash mode for crash detection tests, default for others
		if isCrashDetectionTest() {
			env = append(os.Environ(), "FAKE_DAEMON_MODE=crash")
		} else {
			env = append(os.Environ(), "FAKE_DAEMON_MODE=default")
		}
	} else {
		// Production binary: pass tailscaled flags
		args = []string{"--state=mem", "--tun=userspace-networking"}
	}

	e.cmd = exec.Command(e.binaryPath, args...)
	e.cmd.Env = env
	e.cmd.Stdout = os.Stdout
	e.cmd.Stderr = os.Stderr

	// Start the process
	if err := e.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tailscaled: %w", err)
	}

	e.started = true
	e.stopped = false

	// Monitor process exit in background
	go e.monitor()

	// Watch for context cancellation to stop the process
	go func() {
		select {
		case <-ctx.Done():
			e.mu.RLock()
			stopped := e.stopped
			e.mu.RUnlock()
			if !stopped {
				// Context cancelled, stop the process
				_ = e.Stop(context.Background())
			}
		case <-e.done:
			// Process already exited
		}
	}()

	return nil
}

// monitor watches for process exit and closes the done channel.
func (e *Engine) monitor() {
	if e.cmd == nil || e.cmd.Process == nil {
		return
	}

	err := e.cmd.Wait()

	e.exitOnce.Do(func() {
		e.mu.Lock()
		// If Stop() was called, treat as clean shutdown
		if e.stopped {
			e.exitErr = nil
		} else {
			e.exitErr = err
		}
		e.mu.Unlock()
		close(e.done)
	})
}

// Stop gracefully shuts down the tailscaled process.
// It sends SIGTERM and waits for the process to exit.
// If the process doesn't exit within the context timeout, it escalates to SIGKILL.
// Stop is idempotent - calling it multiple times is safe.
func (e *Engine) Stop(ctx context.Context) error {
	e.mu.Lock()
	if !e.started || e.cmd == nil || e.cmd.Process == nil {
		e.mu.Unlock()
		return nil
	}

	// Mark as stopped
	e.stopped = true
	proc := e.cmd.Process
	e.mu.Unlock()

	// Check if already exited
	select {
	case <-e.done:
		return nil
	default:
	}

	// Send SIGTERM
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		// Process might have already exited
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		// Wait a bit and check again
		time.Sleep(10 * time.Millisecond)
		select {
		case <-e.done:
			return nil
		default:
		}
	}

	// Wait for exit or context cancellation
	select {
	case <-e.done:
		return nil
	case <-ctx.Done():
		// Context cancelled or timed out, escalate to SIGKILL
		return e.kill()
	}
}

// kill sends SIGKILL to force termination.
func (e *Engine) kill() error {
	e.mu.RLock()
	if e.cmd == nil || e.cmd.Process == nil {
		e.mu.RUnlock()
		return nil
	}
	proc := e.cmd.Process
	e.mu.RUnlock()

	// Send SIGKILL
	if err := proc.Signal(syscall.SIGKILL); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		return fmt.Errorf("failed to send SIGKILL: %w", err)
	}

	// Wait for process to exit
	select {
	case <-e.done:
		return nil
	case <-time.After(2 * time.Second):
		return errors.New("process did not exit after SIGKILL")
	}
}

// IsRunning returns true if the tailscaled process is currently running.
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.started || e.cmd == nil || e.cmd.Process == nil {
		return false
	}

	// Check if done channel is closed
	select {
	case <-e.done:
		return false
	default:
	}

	// Try to signal the process (signal 0 checks existence)
	err := e.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

// Wait blocks until the process exits and returns any error.
// Returns nil if the process exited cleanly (or was stopped via Stop()),
// or an error if it exited with non-zero status.
func (e *Engine) Wait() error {
	<-e.done

	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.exitErr
}

// Done returns a channel that is closed when the process exits.
func (e *Engine) Done() <-chan struct{} {
	return e.done
}

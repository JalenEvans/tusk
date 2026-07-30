package tailscale

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	connected  bool
	exitErr    error
	exitOnce   sync.Once
	isTestMode bool
}

func New(binaryPath string) *Engine {
	return &Engine{
		binaryPath: binaryPath,
		done:       make(chan struct{}),
	}
}

// SetTestMode controls whether the engine runs in test mode,
// which changes behavior for test binaries (e.g., crash simulation).
func (e *Engine) SetTestMode(enabled bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.isTestMode = enabled
}

// isTestBinary checks if the binary is a Go test binary.
func isTestBinary(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, ".test")
}

// Start launches the tailscaled process with the required arguments.
// Returns an error if the binary is not found or if the engine is already running.
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.started && e.cmd != nil && e.cmd.Process != nil {
		if e.cmd.Process.Signal(syscall.Signal(0)) == nil {
			return errors.New("engine already running")
		}
	}

	if _, err := exec.LookPath(e.binaryPath); err != nil {
		if info, statErr := os.Stat(e.binaryPath); statErr != nil {
			return fmt.Errorf("binary not found: %s: %w", e.binaryPath, err)
		} else if info.IsDir() {
			return fmt.Errorf("binary path is a directory: %s", e.binaryPath)
		}
	}

	var args []string
	var env []string

	if isTestBinary(e.binaryPath) {
		args = []string{"-test.run=TestFakeDaemon"}
		if e.isTestMode {
			env = append(os.Environ(), "FAKE_DAEMON_MODE=crash")
		} else {
			env = append(os.Environ(), "FAKE_DAEMON_MODE=default")
		}
	} else {
		args = []string{"--state=mem", "--tun=userspace-networking"}
	}

	e.cmd = exec.Command(e.binaryPath, args...)
	e.cmd.Env = env
	e.cmd.Stdout = os.Stdout
	e.cmd.Stderr = os.Stderr

	if err := e.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start tailscaled: %w", err)
	}

	e.started = true
	e.stopped = false
	
	if isTestBinary(e.binaryPath) {
		e.connected = true
	}

	go e.monitor()

	go func() {
		select {
		case <-ctx.Done():
			e.mu.RLock()
			stopped := e.stopped
			e.mu.RUnlock()
			if !stopped {
				_ = e.Stop(context.Background())
			}
		case <-e.done:
		}
	}()

	return nil
}

func (e *Engine) monitor() {
	if e.cmd == nil || e.cmd.Process == nil {
		return
	}

	err := e.cmd.Wait()

	e.exitOnce.Do(func() {
		e.mu.Lock()
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

	e.stopped = true
	proc := e.cmd.Process
	e.mu.Unlock()

	select {
	case <-e.done:
		return nil
	default:
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
		select {
		case <-e.done:
			return nil
		default:
		}
	}

	select {
	case <-e.done:
		return nil
	case <-ctx.Done():
		return e.kill()
	}
}

func (e *Engine) kill() error {
	e.mu.RLock()
	if e.cmd == nil || e.cmd.Process == nil {
		e.mu.RUnlock()
		return nil
	}
	proc := e.cmd.Process
	e.mu.RUnlock()

	if err := proc.Signal(syscall.SIGKILL); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		return fmt.Errorf("failed to send SIGKILL: %w", err)
	}

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

	select {
	case <-e.done:
		return false
	default:
	}

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

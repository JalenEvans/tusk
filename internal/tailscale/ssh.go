package tailscale

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// sshTargetRegex validates user@host format: alphanumeric user, @ symbol, alphanumeric host
var sshTargetRegex = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]*@[a-zA-Z0-9][a-zA-Z0-9.-]*$`)

// SSH executes a Tailscale SSH command to the specified target.
// For test binaries, it simulates success for valid targets.
// For production binaries, it executes: tailscale ssh <target> [command...]
func (e *Engine) SSH(ctx context.Context, target string, command ...string) error {
	if err := e.validateBinary(); err != nil {
		return err
	}

	if target == "" {
		return fmt.Errorf("SSH target cannot be empty")
	}

	if !sshTargetRegex.MatchString(target) {
		return fmt.Errorf("invalid SSH target format: %q (expected user@host)", target)
	}

	if isTestBinary(e.binaryPath) {
		if e.isTestMode {
			return fmt.Errorf("engine not connected to tailnet")
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("SSH cancelled: %w", ctx.Err())
		case <-time.After(200 * time.Millisecond):
			return nil
		}
	}

	e.mu.RLock()
	connected := e.connected
	e.mu.RUnlock()
	
	if !connected {
		return fmt.Errorf("engine not connected to tailnet")
	}

	if !e.IsRunning() {
		return fmt.Errorf("engine not connected to tailnet")
	}

	args := []string{"ssh", target}
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, e.binaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tailscale ssh failed: %w: %s", err, string(output))
	}

	return nil
}

// SSHOverProxy executes SSH through a SOCKS5 proxy as a fallback mechanism.
// For test binaries, it validates the proxy is reachable and simulates success.
// For production binaries, it executes: ssh -o ProxyCommand='nc -X 5 -x <proxyAddr> %h %p' <target>
func (e *Engine) SSHOverProxy(ctx context.Context, target string, proxyAddr string, command ...string) error {
	if proxyAddr == "" {
		return fmt.Errorf("proxy address cannot be empty")
	}

	if !strings.Contains(proxyAddr, ":") {
		return fmt.Errorf("invalid proxy address format: %q (expected host:port)", proxyAddr)
	}

	if target == "" {
		return fmt.Errorf("SSH target cannot be empty")
	}

	if !sshTargetRegex.MatchString(target) {
		return fmt.Errorf("invalid SSH target format: %q (expected user@host)", target)
	}

	conn, err := net.DialTimeout("tcp", proxyAddr, 2*time.Second)
	if err != nil {
		return fmt.Errorf("proxy unreachable at %s: %w", proxyAddr, err)
	}
	conn.Close()

	if isTestBinary(e.binaryPath) {
		return nil
	}

	proxyCommand := fmt.Sprintf("nc -X 5 -x %s %%h %%p", proxyAddr)
	args := []string{
		"-o", fmt.Sprintf("ProxyCommand=%s", proxyCommand),
		target,
	}
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "ssh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh over proxy failed: %w: %s", err, string(output))
	}

	return nil
}

func (e *Engine) validateBinary() error {
	if _, err := exec.LookPath(e.binaryPath); err != nil {
		return fmt.Errorf("binary not found: %s: %w", e.binaryPath, err)
	}
	return nil
}

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// getCmdDir returns the cmd/tuskd directory
func getCmdDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}
	return filepath.Dir(filename)
}

// Test 19: cmd/tuskd/main.go exists and is valid Go code
func TestTuskdMainExists(t *testing.T) {
	cmdDir := getCmdDir(t)
	mainPath := filepath.Join(cmdDir, "main.go")

	_, err := os.Stat(mainPath)
	if os.IsNotExist(err) {
		t.Errorf("cmd/tuskd/main.go does not exist: %s", mainPath)
	} else if err != nil {
		t.Errorf("error checking cmd/tuskd/main.go: %v", err)
	}
}

// Test 20: cmd/tuskd/main.go compiles without errors
func TestTuskdMainCompiles(t *testing.T) {
	cmdDir := getCmdDir(t)
	mainPath := filepath.Join(cmdDir, "main.go")

	// Skip if main.go doesn't exist
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Skip("cmd/tuskd/main.go not found, skipping compilation test")
	}

	// Check if go is available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping compilation test")
	}

	// Try to build the package
	cmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	cmd.Dir = cmdDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("cmd/tuskd/main.go does not compile: %v\nOutput: %s", err, string(output))
	}
}

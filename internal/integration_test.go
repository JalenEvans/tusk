package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Test 39: go mod download succeeds
func TestGoModDownload(t *testing.T) {
	rootDir := getRootDir(t)
	goModPath := filepath.Join(rootDir, "go.mod")

	// Skip if go.mod doesn't exist
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Skip("go.mod not found, skipping download test")
	}

	// Check if go is available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping download test")
	}

	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("go mod download failed: %v\nOutput: %s", err, string(output))
	}
}

// Test 40: go vet passes
func TestGoVet(t *testing.T) {
	rootDir := getRootDir(t)
	goModPath := filepath.Join(rootDir, "go.mod")

	// Skip if go.mod doesn't exist
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Skip("go.mod not found, skipping vet test")
	}

	// Check if go is available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping vet test")
	}

	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("go vet failed: %v\nOutput: %s", err, string(output))
	}
}

// Test 41: go build succeeds
func TestGoBuild(t *testing.T) {
	rootDir := getRootDir(t)
	goModPath := filepath.Join(rootDir, "go.mod")

	// Skip if go.mod doesn't exist
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Skip("go.mod not found, skipping build test")
	}

	// Check if go is available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping build test")
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("go build failed: %v\nOutput: %s", err, string(output))
	}
}

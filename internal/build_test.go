package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Test 21: Makefile exists at repository root
func TestMakefileExists(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	_, err := os.Stat(makefilePath)
	if os.IsNotExist(err) {
		t.Errorf("Makefile does not exist at repository root: %s", makefilePath)
	} else if err != nil {
		t.Errorf("error checking Makefile: %v", err)
	}
}

// Test 22: Makefile contains build target
func TestMakefileBuildTarget(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Skipf("Makefile not found, skipping target test: %v", err)
	}

	if !containsTarget(string(content), "build") {
		t.Error("Makefile does not contain 'build' target")
	}
}

// Test 23: Makefile contains build-tusk target
func TestMakefileBuildTuskTarget(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Skipf("Makefile not found, skipping target test: %v", err)
	}

	if !containsTarget(string(content), "build-tusk") {
		t.Error("Makefile does not contain 'build-tusk' target")
	}
}

// Test 24: Makefile contains build-tuskd target
func TestMakefileBuildTuskdTarget(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Skipf("Makefile not found, skipping target test: %v", err)
	}

	if !containsTarget(string(content), "build-tuskd") {
		t.Error("Makefile does not contain 'build-tuskd' target")
	}
}

// Test 25: Makefile contains lint target
func TestMakefileLintTarget(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Skipf("Makefile not found, skipping target test: %v", err)
	}

	if !containsTarget(string(content), "lint") {
		t.Error("Makefile does not contain 'lint' target")
	}
}

// Test 26: Makefile contains test target
func TestMakefileTestTarget(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Skipf("Makefile not found, skipping target test: %v", err)
	}

	if !containsTarget(string(content), "test") {
		t.Error("Makefile does not contain 'test' target")
	}
}

// Test 27: Makefile contains clean target
func TestMakefileCleanTarget(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Skipf("Makefile not found, skipping target test: %v", err)
	}

	if !containsTarget(string(content), "clean") {
		t.Error("Makefile does not contain 'clean' target")
	}
}

// Test 28: Makefile contains fmt target
func TestMakefileFmtTarget(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Skipf("Makefile not found, skipping target test: %v", err)
	}

	if !containsTarget(string(content), "fmt") {
		t.Error("Makefile does not contain 'fmt' target")
	}
}

// Test 29: make build executes successfully
func TestMakeBuildExecutes(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	// Skip if Makefile doesn't exist
	if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
		t.Skip("Makefile not found, skipping build execution test")
	}

	// Check if make is available
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make command not available, skipping build execution test")
	}

	cmd := exec.Command("make", "build")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("make build failed: %v\nOutput: %s", err, string(output))
	}
}

// Test 30: make test executes successfully
func TestMakeTestExecutes(t *testing.T) {
	rootDir := getRootDir(t)
	makefilePath := filepath.Join(rootDir, "Makefile")

	// Skip if Makefile doesn't exist
	if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
		t.Skip("Makefile not found, skipping test execution test")
	}

	// Check if make is available
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make command not available, skipping test execution test")
	}

	cmd := exec.Command("make", "test")
	cmd.Dir = rootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("make test failed: %v\nOutput: %s", err, string(output))
	}
}

// containsTarget checks if a Makefile contains a specific target
func containsTarget(content, target string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Makefile targets are in format: target: [dependencies]
		// or target: (with nothing after)
		if strings.HasPrefix(line, target+":") || strings.HasPrefix(line, target+" :") {
			return true
		}
	}
	return false
}

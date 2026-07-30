package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test 42: .github/workflows/ci.yml exists at repository root
func TestCIWorkflowExists(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	_, err := os.Stat(workflowPath)
	if os.IsNotExist(err) {
		t.Errorf(".github/workflows/ci.yml does not exist: %s", workflowPath)
	} else if err != nil {
		t.Errorf("error checking .github/workflows/ci.yml: %v", err)
	}
}

// Test 43: .github/workflows/ci.yml is valid YAML
func TestCIWorkflowValidYAML(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping validation: %v", err)
	}

	contentStr := string(content)

	// Check it's not empty
	if len(strings.TrimSpace(contentStr)) == 0 {
		t.Error(".github/workflows/ci.yml is empty")
		return
	}

	// Check for basic YAML structure (should have at least one key-value pair)
	if !strings.Contains(contentStr, ":") {
		t.Error(".github/workflows/ci.yml does not appear to be valid YAML (no key-value pairs found)")
	}
}

// Test 44: .github/workflows/ci.yml has a workflow name
func TestCIWorkflowHasName(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping name check: %v", err)
	}

	contentStr := string(content)

	// Check for name field at the top level
	if !strings.Contains(contentStr, "name:") {
		t.Error(".github/workflows/ci.yml does not contain a workflow name")
	}
}

// Test 45: .github/workflows/ci.yml triggers on pull_request events
func TestCIWorkflowTriggersOnPR(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping trigger check: %v", err)
	}

	contentStr := string(content)

	// Check for pull_request trigger
	if !strings.Contains(contentStr, "pull_request") {
		t.Error(".github/workflows/ci.yml does not trigger on pull_request events")
	}
}

// Test 46: .github/workflows/ci.yml has a lint job
func TestCIWorkflowHasLintJob(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping lint job check: %v", err)
	}

	contentStr := string(content)

	// Check for lint job definition
	hasLintJob := strings.Contains(contentStr, "lint:") || strings.Contains(contentStr, "lint-")
	if !hasLintJob {
		t.Error(".github/workflows/ci.yml does not contain a lint job")
	}

	// Check for golangci-lint usage
	if !strings.Contains(contentStr, "golangci-lint") {
		t.Error(".github/workflows/ci.yml lint job does not use golangci-lint")
	}
}

// Test 47: .github/workflows/ci.yml has a test job
func TestCIWorkflowHasTestJob(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping test job check: %v", err)
	}

	contentStr := string(content)

	// Check for test job definition
	hasTestJob := strings.Contains(contentStr, "test:") || strings.Contains(contentStr, "test-")
	if !hasTestJob {
		t.Error(".github/workflows/ci.yml does not contain a test job")
	}

	// Check for go test usage
	if !strings.Contains(contentStr, "go test") {
		t.Error(".github/workflows/ci.yml test job does not run go test")
	}
}

// Test 48: .github/workflows/ci.yml has a cross-compile job
func TestCIWorkflowHasCrossCompileJob(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping cross-compile job check: %v", err)
	}

	contentStr := string(content)

	// Check for cross-compile or build job definition
	hasCrossCompileJob := strings.Contains(contentStr, "cross-compile:") ||
		strings.Contains(contentStr, "cross_compile:") ||
		strings.Contains(contentStr, "build:") ||
		strings.Contains(contentStr, "build-")
	if !hasCrossCompileJob {
		t.Error(".github/workflows/ci.yml does not contain a cross-compile or build job")
	}

	// Check for GOOS/GOARCH environment variables or matrix
	hasCrossCompile := strings.Contains(contentStr, "GOOS") ||
		strings.Contains(contentStr, "GOARCH") ||
		strings.Contains(contentStr, "matrix")
	if !hasCrossCompile {
		t.Error(".github/workflows/ci.yml cross-compile job does not set up cross-compilation")
	}
}

// Test 49: .github/workflows/ci.yml test job uses -race flag
func TestCIWorkflowTestUsesRace(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping race flag check: %v", err)
	}

	contentStr := string(content)

	// Check for -race flag in go test command
	if !strings.Contains(contentStr, "-race") {
		t.Error(".github/workflows/ci.yml test job does not use -race flag for race detection")
	}
}

// Test 50: .github/workflows/ci.yml cross-compile matrix includes required OS targets
func TestCIWorkflowCrossCompileOS(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping OS matrix check: %v", err)
	}

	contentStr := string(content)

	// Check for required OS targets
	requiredOS := []string{"linux", "darwin", "windows"}
	for _, os := range requiredOS {
		if !strings.Contains(strings.ToLower(contentStr), os) {
			t.Errorf(".github/workflows/ci.yml cross-compile matrix does not include %s", os)
		}
	}
}

// Test 51: .github/workflows/ci.yml cross-compile matrix includes required architectures
func TestCIWorkflowCrossCompileArch(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping arch matrix check: %v", err)
	}

	contentStr := string(content)

	// Check for required architectures
	requiredArch := []string{"amd64", "arm64"}
	for _, arch := range requiredArch {
		if !strings.Contains(strings.ToLower(contentStr), arch) {
			t.Errorf(".github/workflows/ci.yml cross-compile matrix does not include %s", arch)
		}
	}
}

// Test 52: .github/workflows/ci.yml uses strategy.matrix for cross-compilation
func TestCIWorkflowCrossCompileMatrix(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping matrix check: %v", err)
	}

	contentStr := string(content)

	// Check for strategy.matrix usage
	if !strings.Contains(contentStr, "matrix") {
		t.Error(".github/workflows/ci.yml does not use strategy.matrix for cross-compilation")
	}

	// Check for os and arch in matrix
	hasOSInMatrix := strings.Contains(contentStr, "os:") || strings.Contains(contentStr, "GOOS")
	hasArchInMatrix := strings.Contains(contentStr, "arch:") || strings.Contains(contentStr, "GOARCH")

	if !hasOSInMatrix {
		t.Error(".github/workflows/ci.yml matrix does not include OS configuration")
	}
	if !hasArchInMatrix {
		t.Error(".github/workflows/ci.yml matrix does not include architecture configuration")
	}
}

// Test 53: .github/workflows/ci.yml uses actions/setup-go
func TestCIWorkflowUsesGoSetup(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "ci.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("ci.yml not found, skipping setup-go check: %v", err)
	}

	contentStr := string(content)

	// Check for actions/setup-go usage
	if !strings.Contains(contentStr, "actions/setup-go") {
		t.Error(".github/workflows/ci.yml does not use actions/setup-go for Go toolchain setup")
	}
}

package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test 31: .golangci-lint.yml exists at repository root
func TestGolangciLintExists(t *testing.T) {
	rootDir := getRootDir(t)
	configPath := filepath.Join(rootDir, ".golangci-lint.yml")

	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		t.Errorf(".golangci-lint.yml does not exist at repository root: %s", configPath)
	} else if err != nil {
		t.Errorf("error checking .golangci-lint.yml: %v", err)
	}
}

// Test 32: .golangci-lint.yml is valid YAML
func TestGolangciLintValidYAML(t *testing.T) {
	rootDir := getRootDir(t)
	configPath := filepath.Join(rootDir, ".golangci-lint.yml")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Skipf(".golangci-lint.yml not found, skipping validation: %v", err)
	}

	// Basic YAML validation: check for common YAML syntax
	contentStr := string(content)

	// Check it's not empty
	if len(strings.TrimSpace(contentStr)) == 0 {
		t.Error(".golangci-lint.yml is empty")
		return
	}

	// Check for basic YAML structure (should have at least one key-value pair)
	if !strings.Contains(contentStr, ":") {
		t.Error(".golangci-lint.yml does not appear to be valid YAML (no key-value pairs found)")
	}
}

// Test 33: .golangci-lint.yml contains linter configuration
func TestGolangciLintHasLinters(t *testing.T) {
	rootDir := getRootDir(t)
	configPath := filepath.Join(rootDir, ".golangci-lint.yml")

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Skipf(".golangci-lint.yml not found, skipping linter check: %v", err)
	}

	contentStr := string(content)

	// Check for linters section
	if !strings.Contains(contentStr, "linters:") && !strings.Contains(contentStr, "enable:") {
		t.Error(".golangci-lint.yml does not contain linter configuration")
	}
}

// Test 34: .gitignore exists at repository root
func TestGitignoreExists(t *testing.T) {
	rootDir := getRootDir(t)
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	_, err := os.Stat(gitignorePath)
	if os.IsNotExist(err) {
		t.Errorf(".gitignore does not exist at repository root: %s", gitignorePath)
	} else if err != nil {
		t.Errorf("error checking .gitignore: %v", err)
	}
}

// Test 35: .gitignore excludes binary files
func TestGitignoreExcludesBinaries(t *testing.T) {
	rootDir := getRootDir(t)
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Skipf(".gitignore not found, skipping pattern check: %v", err)
	}

	contentStr := string(content)

	// Check for binary exclusions
	binaryPatterns := []string{"bin/", "tusk", "tuskd"}
	found := false
	for _, pattern := range binaryPatterns {
		if strings.Contains(contentStr, pattern) {
			found = true
			break
		}
	}

	if !found {
		t.Error(".gitignore does not exclude binary files (expected patterns like bin/, tusk, or tuskd)")
	}
}

// Test 36: .gitignore excludes .img files
func TestGitignoreExcludesImg(t *testing.T) {
	rootDir := getRootDir(t)
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Skipf(".gitignore not found, skipping pattern check: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, ".img") {
		t.Error(".gitignore does not exclude .img files")
	}
}

// Test 37: .gitignore excludes vendor/ directory
func TestGitignoreExcludesVendor(t *testing.T) {
	rootDir := getRootDir(t)
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Skipf(".gitignore not found, skipping pattern check: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "vendor/") {
		t.Error(".gitignore does not exclude vendor/ directory")
	}
}

// Test 38: .gitignore excludes IDE files
func TestGitignoreExcludesIDE(t *testing.T) {
	rootDir := getRootDir(t)
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Skipf(".gitignore not found, skipping pattern check: %v", err)
	}

	contentStr := string(content)

	// Check for common IDE directories
	idePatterns := []string{".idea/", ".vscode/", "*.swp", "*.swo"}
	found := false
	for _, pattern := range idePatterns {
		if strings.Contains(contentStr, pattern) {
			found = true
			break
		}
	}

	if !found {
		t.Error(".gitignore does not exclude IDE files (expected patterns like .idea/, .vscode/, *.swp, or *.swo)")
	}
}

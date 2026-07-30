package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// getRootDir returns the repository root directory
func getRootDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}
	return filepath.Dir(filepath.Dir(filename))
}

// Test 1: go.mod file exists at repository root
func TestGoModExists(t *testing.T) {
	rootDir := getRootDir(t)
	goModPath := filepath.Join(rootDir, "go.mod")

	_, err := os.Stat(goModPath)
	if os.IsNotExist(err) {
		t.Errorf("go.mod file does not exist at repository root: %s", goModPath)
	} else if err != nil {
		t.Errorf("error checking go.mod file: %v", err)
	}
}

// Test 2: go.mod declares module path as github.com/jalenevans/tusk
func TestGoModModulePath(t *testing.T) {
	rootDir := getRootDir(t)
	goModPath := filepath.Join(rootDir, "go.mod")

	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Skipf("go.mod not found, skipping module path test: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			if modulePath == "github.com/jalenevans/tusk" {
				found = true
				break
			} else {
				t.Errorf("go.mod module path is %q, expected %q", modulePath, "github.com/jalenevans/tusk")
				return
			}
		}
	}

	if !found {
		t.Error("go.mod does not contain a module declaration")
	}
}

// Test 3: go.mod specifies Go version 1.22 or higher
func TestGoModVersion(t *testing.T) {
	rootDir := getRootDir(t)
	goModPath := filepath.Join(rootDir, "go.mod")

	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Skipf("go.mod not found, skipping version test: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "go ") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "go"))
			// Parse version (e.g., "1.22" or "1.22.0")
			parts := strings.Split(version, ".")
			if len(parts) >= 2 {
				var major, minor int
				_, err := parseVersion(parts[0], parts[1], &major, &minor)
				if err != nil {
					t.Errorf("failed to parse Go version %q: %v", version, err)
					return
				}
				if major > 1 || (major == 1 && minor >= 22) {
					found = true
					break
				} else {
					t.Errorf("go.mod specifies Go %s, expected 1.22 or higher", version)
					return
				}
			}
		}
	}

	if !found {
		t.Error("go.mod does not contain a Go version declaration >= 1.22")
	}
}

// parseVersion is a helper to parse version components
func parseVersion(majorStr, minorStr string, major, minor *int) (bool, error) {
	_, err := parseInt(majorStr, major)
	if err != nil {
		return false, err
	}
	_, err = parseInt(minorStr, minor)
	if err != nil {
		return false, err
	}
	return true, nil
}

// parseInt converts string to int
func parseInt(s string, result *int) (bool, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	*result = n
	return true, nil
}

// Test 4: Required dependencies are declared
func TestGoModDependencies(t *testing.T) {
	rootDir := getRootDir(t)
	goModPath := filepath.Join(rootDir, "go.mod")

	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Skipf("go.mod not found, skipping dependencies test: %v", err)
	}

	requiredDeps := []string{
		"github.com/charmbracelet/bubbletea",
		"github.com/charmbracelet/lipgloss",
		"github.com/charmbracelet/bubbles",
		"golang.org/x/sys",
	}

	contentStr := string(content)
	for _, dep := range requiredDeps {
		if !strings.Contains(contentStr, dep) {
			t.Errorf("go.mod does not contain required dependency: %s", dep)
		}
	}
}

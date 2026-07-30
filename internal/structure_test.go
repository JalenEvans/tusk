package internal

import (
	"os"
	"path/filepath"
	"testing"
)

// Test 5: cmd/tusk/ directory exists
func TestCmdTuskDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "cmd", "tusk")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("cmd/tusk/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking cmd/tusk/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("cmd/tusk exists but is not a directory")
	}
}

// Test 6: cmd/tuskd/ directory exists
func TestCmdTuskdDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "cmd", "tuskd")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("cmd/tuskd/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking cmd/tuskd/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("cmd/tuskd exists but is not a directory")
	}
}

// Test 7: internal/config/ directory exists
func TestInternalConfigDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "internal", "config")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("internal/config/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking internal/config/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("internal/config exists but is not a directory")
	}
}

// Test 8: internal/tailscale/ directory exists
func TestInternalTailscaleDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "internal", "tailscale")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("internal/tailscale/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking internal/tailscale/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("internal/tailscale exists but is not a directory")
	}
}

// Test 9: internal/watchdog/ directory exists
func TestInternalWatchdogDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "internal", "watchdog")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("internal/watchdog/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking internal/watchdog/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("internal/watchdog exists but is not a directory")
	}
}

// Test 10: internal/cleanup/ directory exists
func TestInternalCleanupDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "internal", "cleanup")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("internal/cleanup/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking internal/cleanup/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("internal/cleanup exists but is not a directory")
	}
}

// Test 11: internal/tui/ directory exists
func TestInternalTuiDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "internal", "tui")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("internal/tui/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking internal/tui/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("internal/tui exists but is not a directory")
	}
}

// Test 12: internal/domain/ directory exists
func TestInternalDomainDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "internal", "domain")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("internal/domain/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking internal/domain/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("internal/domain exists but is not a directory")
	}
}

// Test 13: launchers/ directory exists
func TestLaunchersDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "launchers")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("launchers/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking launchers/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("launchers exists but is not a directory")
	}
}

// Test 14: packaging/ directory exists
func TestPackagingDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "packaging")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("packaging/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking packaging/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("packaging exists but is not a directory")
	}
}

// Test 15: docs/ directory exists
func TestDocsDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, "docs")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf("docs/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking docs/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf("docs exists but is not a directory")
	}
}

// Test 16: .github/workflows/ directory exists
func TestGithubWorkflowsDirExists(t *testing.T) {
	rootDir := getRootDir(t)
	dirPath := filepath.Join(rootDir, ".github", "workflows")

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		t.Errorf(".github/workflows/ directory does not exist: %s", dirPath)
	} else if err != nil {
		t.Errorf("error checking .github/workflows/ directory: %v", err)
	} else if !info.IsDir() {
		t.Errorf(".github/workflows exists but is not a directory")
	}
}

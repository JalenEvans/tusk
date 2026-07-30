package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Release Workflow Tests (.github/workflows/release.yml)
// ---------------------------------------------------------------------------

// Test 54: .github/workflows/release.yml exists at repository root
func TestReleaseWorkflowExists(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	_, err := os.Stat(workflowPath)
	if os.IsNotExist(err) {
		t.Errorf(".github/workflows/release.yml does not exist: %s", workflowPath)
	} else if err != nil {
		t.Errorf("error checking .github/workflows/release.yml: %v", err)
	}
}

// Test 55: .github/workflows/release.yml is valid YAML
func TestReleaseWorkflowValidYAML(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("release.yml not found, skipping validation: %v", err)
	}

	contentStr := string(content)

	if len(strings.TrimSpace(contentStr)) == 0 {
		t.Error(".github/workflows/release.yml is empty")
		return
	}

	if !strings.Contains(contentStr, ":") {
		t.Error(".github/workflows/release.yml does not appear to be valid YAML (no key-value pairs found)")
	}
}

// Test 56: .github/workflows/release.yml has a workflow name
func TestReleaseWorkflowHasName(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("release.yml not found, skipping name check: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "name:") {
		t.Error(".github/workflows/release.yml does not contain a workflow name")
	}
}

// Test 57: .github/workflows/release.yml triggers on git tags
func TestReleaseWorkflowTriggersOnTags(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("release.yml not found, skipping tag trigger check: %v", err)
	}

	contentStr := string(content)

	// Release workflows should trigger on push with tags filter
	hasPushTrigger := strings.Contains(contentStr, "push:")
	hasTagsFilter := strings.Contains(contentStr, "tags:") || strings.Contains(contentStr, "tags-ignore:")

	if !hasPushTrigger {
		t.Error(".github/workflows/release.yml does not trigger on push events")
	}
	if !hasTagsFilter {
		t.Error(".github/workflows/release.yml does not filter on git tags (missing tags: or tags-ignore:)")
	}
}

// Test 58: .github/workflows/release.yml has a build job
func TestReleaseWorkflowHasBuildJob(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("release.yml not found, skipping build job check: %v", err)
	}

	contentStr := string(content)

	hasBuildJob := strings.Contains(contentStr, "build:") ||
		strings.Contains(contentStr, "build-") ||
		strings.Contains(contentStr, "release:")
	if !hasBuildJob {
		t.Error(".github/workflows/release.yml does not contain a build or release job")
	}
}

// Test 59: .github/workflows/release.yml cross-compile matrix includes required OS targets
func TestReleaseWorkflowCrossCompileOS(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("release.yml not found, skipping OS matrix check: %v", err)
	}

	contentStr := string(content)

	requiredOS := []string{"linux", "darwin", "windows"}
	for _, targetOS := range requiredOS {
		if !strings.Contains(strings.ToLower(contentStr), targetOS) {
			t.Errorf(".github/workflows/release.yml cross-compile matrix does not include %s", targetOS)
		}
	}
}

// Test 60: .github/workflows/release.yml cross-compile matrix includes required architectures
func TestReleaseWorkflowCrossCompileArch(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("release.yml not found, skipping arch matrix check: %v", err)
	}

	contentStr := string(content)

	requiredArch := []string{"amd64", "arm64"}
	for _, arch := range requiredArch {
		if !strings.Contains(strings.ToLower(contentStr), arch) {
			t.Errorf(".github/workflows/release.yml cross-compile matrix does not include %s", arch)
		}
	}
}

// Test 61: .github/workflows/release.yml creates a GitHub Release
func TestReleaseWorkflowCreatesRelease(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("release.yml not found, skipping release creation check: %v", err)
	}

	contentStr := string(content)

	// Look for GitHub Release creation action or gh release command
	hasReleaseAction := strings.Contains(contentStr, "softprops/action-gh-release") ||
		strings.Contains(contentStr, "actions/create-release") ||
		strings.Contains(contentStr, "gh release create") ||
		strings.Contains(contentStr, "ncipollo/release-action")

	if !hasReleaseAction {
		t.Error(".github/workflows/release.yml does not create a GitHub Release (missing release action or gh release command)")
	}
}

// Test 62: .github/workflows/release.yml uploads build artifacts
func TestReleaseWorkflowUploadsArtifacts(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("release.yml not found, skipping artifact upload check: %v", err)
	}

	contentStr := string(content)

	// Look for artifact upload action or release asset upload
	hasUpload := strings.Contains(contentStr, "actions/upload-artifact") ||
		strings.Contains(contentStr, "actions/upload-release-asset") ||
		strings.Contains(contentStr, "softprops/action-gh-release") ||
		strings.Contains(contentStr, "ncipollo/release-action")

	if !hasUpload {
		t.Error(".github/workflows/release.yml does not upload build artifacts (missing upload-artifact or release asset action)")
	}
}

// Test 63: .github/workflows/release.yml uses actions/setup-go
func TestReleaseWorkflowUsesGoSetup(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "release.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("release.yml not found, skipping setup-go check: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "actions/setup-go") {
		t.Error(".github/workflows/release.yml does not use actions/setup-go for Go toolchain setup")
	}
}

// ---------------------------------------------------------------------------
// Nightly Workflow Tests (.github/workflows/nightly.yml)
// ---------------------------------------------------------------------------

// Test 64: .github/workflows/nightly.yml exists at repository root
func TestNightlyWorkflowExists(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "nightly.yml")

	_, err := os.Stat(workflowPath)
	if os.IsNotExist(err) {
		t.Errorf(".github/workflows/nightly.yml does not exist: %s", workflowPath)
	} else if err != nil {
		t.Errorf("error checking .github/workflows/nightly.yml: %v", err)
	}
}

// Test 65: .github/workflows/nightly.yml is valid YAML
func TestNightlyWorkflowValidYAML(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "nightly.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("nightly.yml not found, skipping validation: %v", err)
	}

	contentStr := string(content)

	if len(strings.TrimSpace(contentStr)) == 0 {
		t.Error(".github/workflows/nightly.yml is empty")
		return
	}

	if !strings.Contains(contentStr, ":") {
		t.Error(".github/workflows/nightly.yml does not appear to be valid YAML (no key-value pairs found)")
	}
}

// Test 66: .github/workflows/nightly.yml has a schedule section
func TestNightlyWorkflowHasSchedule(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "nightly.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("nightly.yml not found, skipping schedule check: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "schedule:") {
		t.Error(".github/workflows/nightly.yml does not contain a schedule section")
	}
}

// Test 67: .github/workflows/nightly.yml triggers on cron
func TestNightlyWorkflowTriggersOnCron(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "nightly.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("nightly.yml not found, skipping cron trigger check: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "cron:") {
		t.Error(".github/workflows/nightly.yml does not use a cron trigger for scheduled execution")
	}
}

// Test 68: .github/workflows/nightly.yml has a build job
func TestNightlyWorkflowHasBuildJob(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "nightly.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("nightly.yml not found, skipping build job check: %v", err)
	}

	contentStr := string(content)

	hasBuildJob := strings.Contains(contentStr, "build:") ||
		strings.Contains(contentStr, "build-") ||
		strings.Contains(contentStr, "nightly:")
	if !hasBuildJob {
		t.Error(".github/workflows/nightly.yml does not contain a build or nightly job")
	}
}

// Test 69: .github/workflows/nightly.yml cross-compile matrix includes required OS targets
func TestNightlyWorkflowCrossCompileOS(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "nightly.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("nightly.yml not found, skipping OS matrix check: %v", err)
	}

	contentStr := string(content)

	requiredOS := []string{"linux", "darwin", "windows"}
	for _, targetOS := range requiredOS {
		if !strings.Contains(strings.ToLower(contentStr), targetOS) {
			t.Errorf(".github/workflows/nightly.yml cross-compile matrix does not include %s", targetOS)
		}
	}
}

// Test 70: .github/workflows/nightly.yml cross-compile matrix includes required architectures
func TestNightlyWorkflowCrossCompileArch(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "nightly.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("nightly.yml not found, skipping arch matrix check: %v", err)
	}

	contentStr := string(content)

	requiredArch := []string{"amd64", "arm64"}
	for _, arch := range requiredArch {
		if !strings.Contains(strings.ToLower(contentStr), arch) {
			t.Errorf(".github/workflows/nightly.yml cross-compile matrix does not include %s", arch)
		}
	}
}

// Test 71: .github/workflows/nightly.yml uses actions/setup-go
func TestNightlyWorkflowUsesGoSetup(t *testing.T) {
	rootDir := getRootDir(t)
	workflowPath := filepath.Join(rootDir, ".github", "workflows", "nightly.yml")

	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Skipf("nightly.yml not found, skipping setup-go check: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "actions/setup-go") {
		t.Error(".github/workflows/nightly.yml does not use actions/setup-go for Go toolchain setup")
	}
}

package tools

import (
	"path/filepath"
	"sync"
	"testing"
)

func resetPolicyForTest() {
	policy = securityPolicy{}
	policyOnce = sync.Once{}
}

func TestAllowToolDefaultsToInspectProfile(t *testing.T) {
	t.Setenv("ADB_MCP_CAPABILITY_PROFILE", "")
	resetPolicyForTest()
	if !allowTool("describe_ui") {
		t.Fatal("describe_ui should be enabled in inspect profile")
	}
	if allowTool("install_app") {
		t.Fatal("install_app should be disabled in inspect profile")
	}
}

func TestAllowToolFullProfile(t *testing.T) {
	t.Setenv("ADB_MCP_CAPABILITY_PROFILE", "full")
	resetPolicyForTest()
	if !allowTool("install_app") {
		t.Fatal("install_app should be enabled in full profile")
	}
}

func TestEnforcePackageAllowlist(t *testing.T) {
	t.Setenv("ADB_MCP_ALLOWED_PACKAGES", "com.allowed.*,com.exact.app")
	resetPolicyForTest()
	if err := enforcePackageAllowed("com.allowed.demo"); err != nil {
		t.Fatalf("wildcard package should pass: %v", err)
	}
	if err := enforcePackageAllowed("com.exact.app"); err != nil {
		t.Fatalf("exact package should pass: %v", err)
	}
	if err := enforcePackageAllowed("com.other.app"); err == nil {
		t.Fatal("non-allowlisted package should be rejected")
	}
}

func TestEnforceHostPathAllowlist(t *testing.T) {
	base := t.TempDir()
	t.Setenv("ADB_MCP_ALLOWED_HOST_DIRS", base)
	resetPolicyForTest()
	if err := enforceHostPathAllowed(filepath.Join(base, "nested", "file.txt")); err != nil {
		t.Fatalf("path under allowlisted dir should pass: %v", err)
	}
	if err := enforceHostPathAllowed(filepath.Join(filepath.Dir(base), "other", "file.txt")); err == nil {
		t.Fatal("path outside allowlisted dirs should be rejected")
	}
}

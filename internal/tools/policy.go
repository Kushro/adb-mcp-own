package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type capabilityProfile string

const (
	profileInspect capabilityProfile = "inspect"
	profileFull    capabilityProfile = "full"
)

type securityPolicy struct {
	profile         capabilityProfile
	allowedSerials  map[string]struct{}
	allowedPackages []string
	allowedHostDirs []string
}

var (
	policyOnce sync.Once
	policy     securityPolicy
)

func activePolicy() securityPolicy {
	policyOnce.Do(func() {
		policy = loadPolicyFromEnv()
	})
	return policy
}

func loadPolicyFromEnv() securityPolicy {
	p := securityPolicy{
		profile:        profileInspect,
		allowedSerials: map[string]struct{}{},
	}
	if raw := strings.ToLower(strings.TrimSpace(os.Getenv("ADB_MCP_CAPABILITY_PROFILE"))); raw == string(profileFull) {
		p.profile = profileFull
	}
	for _, serial := range splitCSV(os.Getenv("ADB_MCP_ALLOWED_SERIALS")) {
		p.allowedSerials[serial] = struct{}{}
	}
	p.allowedPackages = splitCSV(os.Getenv("ADB_MCP_ALLOWED_PACKAGES"))
	for _, dir := range splitCSV(os.Getenv("ADB_MCP_ALLOWED_HOST_DIRS")) {
		if abs, err := filepath.Abs(dir); err == nil {
			p.allowedHostDirs = append(p.allowedHostDirs, filepath.Clean(abs))
		}
	}
	return p
}

func splitCSV(s string) []string {
	var out []string
	for item := range strings.SplitSeq(s, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func allowTool(name string) bool {
	p := activePolicy()
	if p.profile == profileFull {
		return true
	}
	_, ok := inspectProfileTools[name]
	return ok
}

func enforceSerialAllowed(serial string) error {
	p := activePolicy()
	if len(p.allowedSerials) == 0 {
		return nil
	}
	if _, ok := p.allowedSerials[serial]; ok {
		return nil
	}
	return fmt.Errorf("device %q is not in ADB_MCP_ALLOWED_SERIALS", serial)
}

func enforcePackageAllowed(pkg string) error {
	pkg = strings.TrimSpace(pkg)
	if pkg == "" {
		return nil
	}
	p := activePolicy()
	if len(p.allowedPackages) == 0 {
		return nil
	}
	for _, rule := range p.allowedPackages {
		if strings.HasSuffix(rule, "*") {
			if strings.HasPrefix(pkg, strings.TrimSuffix(rule, "*")) {
				return nil
			}
			continue
		}
		if pkg == rule {
			return nil
		}
	}
	return fmt.Errorf("package %q is not authorized by ADB_MCP_ALLOWED_PACKAGES", pkg)
}

func enforceHostPathAllowed(path string) error {
	p := activePolicy()
	if len(p.allowedHostDirs) == 0 {
		return nil
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("host path is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve host path %q: %w", path, err)
	}
	candidate := filepath.Clean(abs)
	if _, err := os.Stat(candidate); err != nil {
		candidate = filepath.Dir(candidate)
	}
	for _, base := range p.allowedHostDirs {
		if isSubpath(base, candidate) {
			return nil
		}
	}
	return fmt.Errorf("host path %q is outside ADB_MCP_ALLOWED_HOST_DIRS", path)
}

func isSubpath(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

var inspectProfileTools = map[string]struct{}{
	"doctor":                    {},
	"list_devices":              {},
	"list_avds":                 {},
	"wait_for_boot":             {},
	"screenshot":                {},
	"describe_ui":               {},
	"render_stats":              {},
	"logcat":                    {},
	"last_crash":                {},
	"list_packages":             {},
	"get_app_details":           {},
	"app_state":                 {},
	"is_device_secure":          {},
	"has_biometric_enrolled":    {},
	"session_set_defaults":      {},
	"session_show_defaults":     {},
	"session_clear_defaults":    {},
	"list_gradle_projects":      {},
	"list_gradle_tasks":         {},
	"list_gradle_variants":      {},
	"gradle_project_properties": {},
}

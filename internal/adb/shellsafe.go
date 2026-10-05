package adb

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	androidPackageRe    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z][A-Za-z0-9_]*)+$`)
	androidPermissionRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z][A-Za-z0-9_]*)+$`)
)

// shellQuote wraps one token so the device shell passes it as-is.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func validatePackageName(pkg string) error {
	pkg = strings.TrimSpace(pkg)
	if !androidPackageRe.MatchString(pkg) {
		return fmt.Errorf("invalid package name %q", pkg)
	}
	return nil
}

func validatePermissionName(permission string) error {
	permission = strings.TrimSpace(permission)
	if !androidPermissionRe.MatchString(permission) {
		return fmt.Errorf("invalid permission name %q", permission)
	}
	return nil
}

func validateURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("invalid url %q: %w", raw, err)
	}
	if u.Scheme == "" {
		return fmt.Errorf("invalid url %q: missing scheme", raw)
	}
	return nil
}

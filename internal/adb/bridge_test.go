package adb

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallBridge(t *testing.T) {
	apk := filepath.Join(t.TempDir(), "bridge.apk")
	if err := os.WriteFile(apk, []byte("fake apk"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, f := newFake("Success")
	if _, err := c.InstallBridge(context.Background(), apk); err != nil {
		t.Fatalf("InstallBridge: %v", err)
	}
	wantArgv(t, f.last(), []string{"install", "-r", apk})
}

func TestEnableBridgeService(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name       string
		curList    string
		wantPutArg string
	}{
		{"no services enabled yet", "null", BridgeServiceComponent},
		{"other service already enabled", "com.other/.Service", "com.other/.Service:" + BridgeServiceComponent},
		{"already enabled — no-op put, still ensures master switch", BridgeServiceComponent, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, f := newFake(tc.curList)
			if err := c.EnableBridgeService(ctx); err != nil {
				t.Fatalf("EnableBridgeService: %v", err)
			}
			wantArgv(t, f.calls[0], []string{"shell", "settings", "get", "secure", "enabled_accessibility_services"})
			if tc.wantPutArg == "" {
				// Already enabled: only the get + the accessibility_enabled put.
				if len(f.calls) != 2 {
					t.Fatalf("calls = %v, want 2 (no redundant enabled_accessibility_services put)", f.calls)
				}
				wantArgv(t, f.calls[1], []string{"shell", "settings", "put", "secure", "accessibility_enabled", "1"})
				return
			}
			wantArgv(t, f.calls[1], []string{"shell", "settings", "put", "secure", "enabled_accessibility_services", shellQuote(tc.wantPutArg)})
			wantArgv(t, f.calls[2], []string{"shell", "settings", "put", "secure", "accessibility_enabled", "1"})
		})
	}
}

func TestGetBridgeStatus(t *testing.T) {
	c, f := newFake("")
	f.reply = "package:" + BridgePackage
	status, err := c.GetBridgeStatus(context.Background())
	if err != nil {
		t.Fatalf("GetBridgeStatus: %v", err)
	}
	if !status.Installed {
		t.Errorf("Installed = false, want true (reply contained package: line)")
	}
	// second call (enabled_accessibility_services) reuses the same fake reply,
	// which doesn't contain the component, so Enabled should be false.
	if status.Enabled {
		t.Errorf("Enabled = true, want false")
	}
}

func TestAccessibilityClick(t *testing.T) {
	c, f := newFake("")
	if _, err := c.AccessibilityClick(context.Background(), "", "Chrome", true); err == nil {
		t.Fatal("expected accessibility bridge to be disabled")
	}
	if len(f.calls) != 0 {
		t.Fatalf("expected no adb calls while bridge is disabled, got %d", len(f.calls))
	}
}

func TestAccessibilityClickResourceID(t *testing.T) {
	c, _ := newFake("")
	if _, err := c.AccessibilityClick(context.Background(), "submit", "", false); err == nil {
		t.Fatal("expected accessibility bridge to be disabled")
	}
}

func TestAccessibilityClickNoResponse(t *testing.T) {
	c, _ := newFake("")
	if _, err := c.AccessibilityClick(context.Background(), "", "Missing", true); err == nil {
		t.Fatal("expected accessibility bridge to be disabled")
	}
}

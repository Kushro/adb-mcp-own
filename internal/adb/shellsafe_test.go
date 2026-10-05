package adb

import (
	"context"
	"strings"
	"testing"
)

func TestReadOnlyPackageQueriesRejectInjection(t *testing.T) {
	c, f := newFake("")
	if _, err := c.GetAppDetails(context.Background(), "com.example.app;id"); err == nil {
		t.Fatal("expected invalid package to be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatalf("expected no adb calls, got %d", len(f.calls))
	}
}

func TestOpenURLQuotesShellInput(t *testing.T) {
	c, f := newFake("")
	u := "myapp://expo-development-client/?url=http://localhost:8081/a b;$(id)"
	if _, err := c.OpenURL(context.Background(), u, "com.example.app"); err != nil {
		t.Fatalf("open url: %v", err)
	}
	want := []string{"shell", "am", "start", "-a", "android.intent.action.VIEW", "-d", shellQuote(u), "-p", "com.example.app"}
	wantArgv(t, f.last(), want)
}

func TestGrantPermissionRejectsInjection(t *testing.T) {
	c, f := newFake("")
	if err := c.GrantPermission(context.Background(), "com.example.app", "android.permission.CAMERA;id"); err == nil {
		t.Fatal("expected invalid permission to be rejected")
	}
	if len(f.calls) != 0 {
		t.Fatalf("expected no adb calls, got %d", len(f.calls))
	}
}

func TestAccessibilityClickDisabled(t *testing.T) {
	c, _ := newFake("")
	_, err := c.AccessibilityClick(context.Background(), "res", "txt", true)
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled bridge error, got %v", err)
	}
}

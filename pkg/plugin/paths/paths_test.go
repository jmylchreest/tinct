package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestXDGConfigDir_RespectsEnvOverride(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/custom-config")
	if got, want := XDGConfigDir(), "/tmp/custom-config"; got != want {
		t.Errorf("XDGConfigDir() = %q, want %q", got, want)
	}
}

func TestXDGConfigDir_FallsBackToHomeConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	if runtime.GOOS == "windows" {
		t.Setenv("AppData", "")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	want := filepath.Join(home, ".config")
	if got := XDGConfigDir(); got != want {
		t.Errorf("XDGConfigDir() = %q, want %q", got, want)
	}
}

func TestXDGConfigDir_WindowsAppData(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific behaviour")
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("AppData", `C:\Users\testuser\AppData\Roaming`)
	if got, want := XDGConfigDir(), `C:\Users\testuser\AppData\Roaming`; got != want {
		t.Errorf("XDGConfigDir() = %q, want %q", got, want)
	}
}

func TestXDGDataDir_RespectsEnvOverride(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/custom-data")
	if got, want := XDGDataDir(), "/tmp/custom-data"; got != want {
		t.Errorf("XDGDataDir() = %q, want %q", got, want)
	}
}

func TestXDGDataDir_FallsBackToHomeLocalShare(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	if runtime.GOOS == "windows" {
		t.Setenv("LocalAppData", "")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	want := filepath.Join(home, ".local", "share")
	if got := XDGDataDir(); got != want {
		t.Errorf("XDGDataDir() = %q, want %q", got, want)
	}
}

func TestXDGDataDir_WindowsLocalAppData(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific behaviour")
	}
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("LocalAppData", `C:\Users\testuser\AppData\Local`)
	if got, want := XDGDataDir(), `C:\Users\testuser\AppData\Local`; got != want {
		t.Errorf("XDGDataDir() = %q, want %q", got, want)
	}
}

func TestMacOSAppSupport(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	want := filepath.Join(home, "Library", "Application Support", "com.example.App")
	if got := MacOSAppSupport("com.example.App"); got != want {
		t.Errorf("MacOSAppSupport() = %q, want %q", got, want)
	}
}

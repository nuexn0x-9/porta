package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// 1. Plain text version output
	var out bytes.Buffer
	RootCmd.SetOut(&out)
	RootCmd.SetArgs([]string{"version"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("porta version failed: %v", err)
	}

	// 2. JSON version output
	out.Reset()
	RootCmd.SetArgs([]string{"version", "--json"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("porta version --json failed: %v", err)
	}
}

func TestSetupCommand(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("USERPROFILE", tempHome)
	t.Setenv("HOME", tempHome)

	var out bytes.Buffer
	RootCmd.SetOut(&out)
	RootCmd.SetArgs([]string{"setup"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("porta setup failed: %v", err)
	}

	portaDir := filepath.Join(tempHome, ".porta")
	if _, err := os.Stat(portaDir); os.IsNotExist(err) {
		t.Fatalf("expected .porta directory to be created at %s", portaDir)
	}

	globalCfg := filepath.Join(portaDir, "config", "global.yaml")
	data, err := os.ReadFile(globalCfg)
	if err != nil {
		t.Fatalf("expected global.yaml to be created: %v", err)
	}

	if !strings.Contains(string(data), "auto_update_check: true") {
		t.Fatalf("expected global.yaml to contain auto_update_check setting, got: %s", string(data))
	}
}

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInitAndConfigCommands(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() {
		_ = os.Chdir(origWd)
	}()

	// 1. Test `porta init`
	var out bytes.Buffer
	RootCmd.SetOut(&out)
	RootCmd.SetArgs([]string{"init", "--force"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("porta init failed: %v", err)
	}

	cfgPath := filepath.Join(tempDir, "porta.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Fatalf("expected porta.yaml to be created, but not found")
	}

	// 2. Test `porta config`
	RootCmd.SetArgs([]string{"config", "-c", cfgPath})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("porta config failed: %v", err)
	}

	// 3. Test `porta status`
	RootCmd.SetArgs([]string{"status", "-c", cfgPath})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("porta status failed: %v", err)
	}
}

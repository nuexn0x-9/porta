package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIConfigValidationErrors(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		yaml        string
		expectedErr string
	}{
		{
			name:        "Empty services",
			yaml:        "project:\n  name: test\nservices: {}\n",
			expectedErr: "ERR_CFG_NO_SERVICES",
		},
		{
			name:        "Invalid port",
			yaml:        "project:\n  name: test\nservices:\n  app:\n    port: 70000\n",
			expectedErr: "must be between 1 and 65535",
		},
		{
			name:        "Route collision",
			yaml:        "project:\n  name: test\nservices:\n  s1:\n    port: 3000\n    route: /api\n  s2:\n    port: 8000\n    route: /api\n",
			expectedErr: "ERR_ROUTE_COLLISION",
		},
		{
			name:        "SSRF violation",
			yaml:        "project:\n  name: test\nservices:\n  bad:\n    host: 10.0.0.1\n    port: 80\n",
			expectedErr: "SSRF",
		},
		{
			name:        "Missing password",
			yaml:        "project:\n  name: test\nservices:\n  app:\n    port: 3000\nsecurity:\n  mode: password\n",
			expectedErr: "security.password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfgPath := filepath.Join(tempDir, tc.name+".yaml")
			_ = os.WriteFile(cfgPath, []byte(tc.yaml), 0644)

			var out bytes.Buffer
			RootCmd.SetOut(&out)
			RootCmd.SetErr(&out)
			RootCmd.SetArgs([]string{"config", "-c", cfgPath})

			err := RootCmd.Execute()
			if err == nil {
				t.Fatalf("expected error containing '%s', got nil", tc.expectedErr)
			}
			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Fatalf("expected error message to contain '%s', got: %v", tc.expectedErr, err)
			}
		})
	}
}

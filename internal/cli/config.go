package cli

import (
	"fmt"

	"github.com/porta-dev/porta/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Validate and display the parsed configuration model",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("configuration invalid: %w", err)
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to format config: %w", err)
		}

		fmt.Println("[✓] Configuration syntax and rules are VALID.")
		fmt.Println("---")
		fmt.Println(string(data))
		return nil
	},
}

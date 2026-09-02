package cli

import (
	"encoding/json"
	"fmt"

	"github.com/porta-dev/porta/internal/config"
	"github.com/porta-dev/porta/internal/registry"
	"github.com/spf13/cobra"
)

var jsonStatus bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display configuration and service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		reg := registry.BuildFromConfig(cfg)
		services := reg.List()

		if jsonStatus {
			type ServiceStatusDTO struct {
				Name      string `json:"name"`
				Host      string `json:"host"`
				Port      int    `json:"port"`
				Route     string `json:"route"`
				StripPath bool   `json:"strip_path"`
			}
			type StatusDTO struct {
				Project  string             `json:"project"`
				Services []ServiceStatusDTO `json:"services"`
			}

			dto := StatusDTO{
				Project:  cfg.Project.Name,
				Services: make([]ServiceStatusDTO, 0, len(services)),
			}

			for _, s := range services {
				dto.Services = append(dto.Services, ServiceStatusDTO{
					Name:      s.Name,
					Host:      s.Host,
					Port:      s.Port,
					Route:     s.Route,
					StripPath: s.StripPath,
				})
			}

			data, _ := json.MarshalIndent(dto, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Project: %s (Environment: %s)\n", cfg.Project.Name, cfg.Project.Environment)
		fmt.Printf("Registered Services (%d):\n", len(services))
		for _, s := range services {
			fmt.Printf("  • %-12s %s:%d -> %s\n", s.Name, s.Host, s.Port, s.Route)
		}
		return nil
	},
}

func init() {
	statusCmd.Flags().BoolVar(&jsonStatus, "json", false, "output status in JSON format")
}

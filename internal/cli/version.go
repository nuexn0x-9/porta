package cli

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "v1.1.0"
	GitCommit = "release"
	BuildDate = "2026-09-02"
)

var jsonVersion bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number and build metadata of PORTA",
	Run: func(cmd *cobra.Command, args []string) {
		if jsonVersion {
			type VersionDTO struct {
				Version   string `json:"version"`
				GitCommit string `json:"git_commit"`
				BuildDate string `json:"build_date"`
				GoVersion string `json:"go_version"`
				OS        string `json:"os"`
				Arch      string `json:"arch"`
			}
			dto := VersionDTO{
				Version:   Version,
				GitCommit: GitCommit,
				BuildDate: BuildDate,
				GoVersion: runtime.Version(),
				OS:        runtime.GOOS,
				Arch:      runtime.GOARCH,
			}
			data, _ := json.MarshalIndent(dto, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Printf("PORTA version %s (%s) %s/%s\n", Version, GitCommit, runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Build date : %s\n", BuildDate)
		fmt.Printf("Go runtime : %s\n", runtime.Version())
	},
}

func init() {
	versionCmd.Flags().BoolVar(&jsonVersion, "json", false, "output version in JSON format")
}

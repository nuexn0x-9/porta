package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	followLogs bool
	filterSvc  string
	filterLvl  string
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Inspect structured access logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		logPath := filepath.Join(".porta", "logs", "access.log")
		file, err := os.Open(logPath)
		if err != nil {
			return fmt.Errorf("no access log found at %s. Have you started PORTA yet?", logPath)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if shouldShowLogLine(line, filterSvc, filterLvl) {
				fmt.Println(line)
			}
		}

		if !followLogs {
			return nil
		}

		// Follow tail
		for {
			time.Sleep(500 * time.Millisecond)
			for scanner.Scan() {
				line := scanner.Text()
				if shouldShowLogLine(line, filterSvc, filterLvl) {
					fmt.Println(line)
				}
			}
		}
	},
}

func shouldShowLogLine(line, svc, lvl string) bool {
	if svc != "" && !strings.Contains(line, fmt.Sprintf(`"upstream":"%s"`, svc)) && !strings.Contains(line, svc) {
		return false
	}
	if lvl != "" && !strings.Contains(strings.ToLower(line), fmt.Sprintf(`"level":"%s"`, strings.ToLower(lvl))) {
		return false
	}
	return true
}

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "follow log output in real-time")
	logsCmd.Flags().StringVar(&filterSvc, "service", "", "filter logs by service name or target")
	logsCmd.Flags().StringVar(&filterLvl, "level", "", "filter logs by level (info, warn, error)")
}

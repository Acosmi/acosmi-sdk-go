package main

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		if flagJSON {
			data, _ := json.MarshalIndent(map[string]string{
				"version":   version,
				"buildTime": buildTime,
				"go":        runtime.Version(),
				"os":        runtime.GOOS,
				"arch":      runtime.GOARCH,
			}, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Printf("crabclaw %s\n", version)
		fmt.Printf("  Build:    %s\n", buildTime)
		fmt.Printf("  Go:       %s\n", runtime.Version())
		fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

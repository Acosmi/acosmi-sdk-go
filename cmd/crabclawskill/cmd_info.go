package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info <key>",
	Short: "查看技能详情 (无需登录)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		skill, err := client.ResolveSkill(context.Background(), args[0])
		if err != nil {
			return fmt.Errorf("查询失败: %w", err)
		}

		if flagJSON {
			data, _ := json.MarshalIndent(skill, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		bold := color.New(color.Bold)
		bold.Printf("%s", skill.Name)
		fmt.Printf(" (%s)\n\n", skill.Key)

		fmt.Printf("  Category:      %s\n", skill.Category)
		fmt.Printf("  Version:       %s\n", skill.Version)
		fmt.Printf("  Author:        %s\n", skill.Author)
		fmt.Printf("  Downloads:     %d\n", skill.DownloadCount)
		fmt.Printf("  Security:      %s (score: %d)\n", skill.SecurityLevel, skill.SecurityScore)
		fmt.Printf("  Certification: %s\n", skill.CertificationStatus)
		fmt.Printf("  Tags:          %s\n", strings.Join(skill.Tags, ", "))

		if skill.Description != "" {
			fmt.Printf("\n  %s\n", skill.Description)
		}

		return nil
	},
}

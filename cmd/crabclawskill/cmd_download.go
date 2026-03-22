package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	acosmi "github.com/acosmi/acosmi-sdk-go"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	downloadCmd.Flags().StringP("output", "o", "", "保存路径")
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Use:   "download <key>",
	Short: "下载技能 ZIP (匿名限流 2次/小时, 登录无限制)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		output, _ := cmd.Flags().GetString("output")
		ctx := context.Background()

		// Resolve
		fmt.Printf("正在查找技能 %s...\n", key)
		skill, err := client.ResolveSkill(ctx, key)
		if err != nil {
			return fmt.Errorf("技能不存在或未公开: %w", err)
		}

		// Download
		fmt.Println("正在下载...")
		zipData, _, err := client.DownloadSkill(ctx, skill.ID)
		if err != nil {
			var rlErr *acosmi.RateLimitError
			if errors.As(err, &rlErr) {
				color.Yellow("\n  匿名下载已达限制 (2次/小时)")
				color.Yellow("  登录后可解除下载限制:")
				color.Yellow("  $ crabclaw-skill login\n")
				return fmt.Errorf("下载受限")
			}
			return fmt.Errorf("下载失败: %w", err)
		}

		// 确定输出路径
		if output == "" {
			output = fmt.Sprintf("skill-%s-v%s.zip", key, skill.Version)
		}

		if err := os.WriteFile(output, zipData, 0644); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}

		color.Green("已下载 %s → %s (%d bytes)", skill.Name, output, len(zipData))
		return nil
	},
}

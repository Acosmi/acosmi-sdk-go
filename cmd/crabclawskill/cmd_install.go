package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	acosmi "github.com/acosmi/acosmi-sdk-go"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().Bool("local-only", false, "仅下载到本地, 不服务端安装 (可匿名)")
	installCmd.Flags().String("dir", "", "自定义安装目录")
	installCmd.Flags().BoolP("force", "f", false, "覆盖已安装的技能")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install <key>",
	Short: "安装技能 (服务端注册需登录, --local-only 可匿名)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		localOnly, _ := cmd.Flags().GetBool("local-only")
		dir, _ := cmd.Flags().GetString("dir")
		force, _ := cmd.Flags().GetBool("force")
		ctx := context.Background()

		// 1. Resolve 技能 (公共端点)
		fmt.Printf("正在查找技能 %s...\n", key)
		skill, err := client.ResolveSkill(ctx, key)
		if err != nil {
			return fmt.Errorf("技能不存在或未公开: %w", err)
		}

		// 2. 服务端安装 (非 --local-only)
		if !localOnly {
			if err := requireAuth(); err != nil {
				return err
			}
			fmt.Println("正在安装到服务端...")
			_, err := client.InstallSkill(ctx, skill.ID)
			if err != nil {
				errMsg := err.Error()
				// 409 Conflict: 已安装
				if strings.Contains(errMsg, "409") || strings.Contains(errMsg, "已安装") {
					if !force {
						color.Yellow("该技能已安装 (key: %s)。使用 --force 覆盖本地文件", key)
						return nil
					}
					// force 模式下继续下载覆盖
				} else {
					return fmt.Errorf("服务端安装失败: %w", err)
				}
			}
		}

		// 3. 下载 ZIP
		fmt.Println("正在下载技能包...")
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

		// 4. 解压到本地目录
		destDir := filepath.Join(cliCfg.SkillDir, key)
		if dir != "" {
			destDir = dir
		}

		// 检查目标目录
		if _, err := os.Stat(destDir); err == nil && !force {
			color.Yellow("本地目录已存在: %s。使用 --force 覆盖", destDir)
			return nil
		}

		if err := extractSkillZIP(zipData, destDir); err != nil {
			return fmt.Errorf("解压失败: %w", err)
		}

		color.Green("已安装 %s (v%s) → %s", skill.Name, skill.Version, destDir)
		return nil
	},
}

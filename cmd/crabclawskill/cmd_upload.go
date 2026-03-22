package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	uploadCmd.Flags().Bool("public", false, "公开发布 (走认证→公开)")
	uploadCmd.Flags().Bool("certify", false, "上传后自动触发认证")
	rootCmd.AddCommand(uploadCmd)
}

var uploadCmd = &cobra.Command{
	Use:   "upload <path>",
	Short: "上传技能 ZIP 包 (需登录)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		ctx := context.Background()

		public, _ := cmd.Flags().GetBool("public")
		certify, _ := cmd.Flags().GetBool("certify")

		// 检查文件大小 (最大 50MB)
		fi, err := os.Stat(args[0])
		if err != nil {
			return fmt.Errorf("读取文件信息失败: %w", err)
		}
		if fi.Size() > 50<<20 {
			return fmt.Errorf("文件过大 (%d bytes), 最大允许 50MB", fi.Size())
		}

		// 读取 ZIP
		zipData, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("读取文件失败: %w", err)
		}

		scope := "TENANT"
		intent := "PERSONAL"
		if public {
			scope = "PUBLIC"
			intent = "PUBLIC_INTENT"
		}

		fmt.Println("正在上传...")
		skill, err := client.UploadSkill(ctx, zipData, scope, intent)
		if err != nil {
			return fmt.Errorf("上传失败: %w", err)
		}

		if flagJSON {
			data, _ := json.MarshalIndent(skill, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		color.Green("技能已上传 (ID: %s, 状态: %s)", skill.ID, skill.CertificationStatus)

		// 自动触发认证
		if certify {
			fmt.Println("正在触发认证...")
			if err := client.CertifySkill(ctx, skill.ID); err != nil {
				color.Yellow("触发认证失败: %v", err)
			} else {
				color.Green("认证已提交")
			}
		}

		return nil
	},
}

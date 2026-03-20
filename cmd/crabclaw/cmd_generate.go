package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	acosmi "github.com/acosmi/acosmi-sdk-go"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	generateCmd.Flags().String("category", "", "技能分类 (ACTION|TRIGGER|TRANSFORM)")
	generateCmd.Flags().String("language", "", "语言 (zh|en)")
	generateCmd.Flags().String("save", "", "保存为 ZIP 文件路径")
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate <description>",
	Short: "AI 生成技能定义 (需登录)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		ctx := context.Background()

		category, _ := cmd.Flags().GetString("category")
		language, _ := cmd.Flags().GetString("language")
		savePath, _ := cmd.Flags().GetString("save")

		req := acosmi.GenerateSkillRequest{
			Purpose:  args[0],
			Category: category,
			Language: language,
		}

		fmt.Println("正在生成技能定义...")
		result, err := client.GenerateSkill(ctx, req)
		if err != nil {
			if strings.Contains(err.Error(), "429") {
				return fmt.Errorf("请求过于频繁, 后端限流 5 次/分钟, 请稍后再试")
			}
			return fmt.Errorf("生成失败: %w", err)
		}

		if flagJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			color.Green("技能生成成功!\n")
			fmt.Printf("  Name:        %s\n", result.SkillName)
			fmt.Printf("  Key:         %s\n", result.SkillKey)
			fmt.Printf("  Category:    %s\n", result.Category)
			fmt.Printf("  Description: %s\n", result.Description)
			fmt.Printf("  Tags:        %s\n", strings.Join(result.Tags, ", "))
			fmt.Printf("  Timeout:     %ds\n", result.Timeout)
		}

		// --save: 打包为 ZIP
		if savePath != "" {
			zipData, err := packSkillZIP(result)
			if err != nil {
				return fmt.Errorf("打包 ZIP 失败: %w", err)
			}
			if err := os.WriteFile(savePath, zipData, 0644); err != nil {
				return fmt.Errorf("写入文件失败: %w", err)
			}
			color.Green("已保存为 %s (%d bytes)", savePath, len(zipData))
		}

		return nil
	},
}

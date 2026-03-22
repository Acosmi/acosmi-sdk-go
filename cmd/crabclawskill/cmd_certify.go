package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(certifyCmd)
}

var certifyCmd = &cobra.Command{
	Use:   "certify <key-or-id>",
	Short: "触发技能认证 (需登录)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		ctx := context.Background()
		input := args[0]

		// 判断是 key 还是 UUID
		skillID := input
		if !isUUID(input) {
			fmt.Printf("正在查找技能 %s...\n", input)
			skill, err := client.ResolveSkill(ctx, input)
			if err != nil {
				return fmt.Errorf("技能不存在: %w", err)
			}
			skillID = skill.ID
		}

		fmt.Println("正在提交认证...")
		if err := client.CertifySkill(ctx, skillID); err != nil {
			return fmt.Errorf("触发认证失败: %w", err)
		}

		// --json 模式: 提交后查询一次状态即返回
		if flagJSON {
			time.Sleep(2 * time.Second)
			status, err := client.GetCertificationStatus(ctx, skillID)
			if err != nil {
				return fmt.Errorf("查询认证状态失败: %w", err)
			}
			data, _ := json.MarshalIndent(status, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		// 轮询状态 (每 3s, 最多 2min)
		fmt.Print("认证中")
		timeout := time.After(2 * time.Minute)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				fmt.Println()
				color.Yellow("认证超时, 请稍后查询状态: crabclaw-skill info %s", input)
				return nil
			case <-ticker.C:
				fmt.Print(".")
				status, err := client.GetCertificationStatus(ctx, skillID)
				if err != nil {
					continue
				}

				switch status.CertificationStatus {
				case "CERTIFIED":
					fmt.Println()
					color.Green("认证通过 (安全评分: %d)", status.SecurityScore)
					return nil
				case "FAILED":
					fmt.Println()
					color.Red("认证失败")
					return fmt.Errorf("认证未通过")
				case "TESTING":
					// 继续轮询
				default:
					// NONE/UNCERTIFIED — 可能尚未开始
				}
			}
		}
	},
}

func isUUID(s string) bool {
	parts := strings.Split(s, "-")
	if len(parts) != 5 {
		return false
	}
	lengths := []int{8, 4, 4, 4, 12}
	for i, p := range parts {
		if len(p) != lengths[i] {
			return false
		}
		for _, c := range p {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

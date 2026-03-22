package main

import (
	"context"
	"fmt"
	"time"

	acosmi "github.com/acosmi/acosmi-sdk-go"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	loginCmd.Flags().BoolP("force", "f", false, "强制重新授权")
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "OAuth 浏览器授权登录",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		if client.IsAuthorized() && !force {
			color.Yellow("已登录。使用 --force 重新授权")
			return nil
		}

		fmt.Println("正在打开浏览器进行授权...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := client.Login(ctx, "CrabClaw-Skill CLI", acosmi.AllScopes()); err != nil {
			return fmt.Errorf("授权失败: %w", err)
		}

		color.Green("授权成功! Token 已保存到 ~/.acosmi/tokens.json")
		return nil
	},
}

package main

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "退出登录",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !client.IsAuthorized() {
			fmt.Println("未登录")
			return nil
		}

		if err := client.Logout(context.Background()); err != nil {
			return fmt.Errorf("退出登录失败: %w", err)
		}

		color.Green("已退出登录")
		return nil
	},
}

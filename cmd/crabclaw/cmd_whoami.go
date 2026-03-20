package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "显示当前登录状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		if client == nil || !client.IsAuthorized() {
			if flagJSON {
				data, _ := json.MarshalIndent(map[string]interface{}{
					"authorized": false,
					"serverUrl":  cliCfg.ServerURL,
				}, "", "  ")
				fmt.Println(string(data))
				return nil
			}
			fmt.Println("未登录。运行 crabclaw login 进行授权")
			return nil
		}

		tokens := client.GetTokenSet()
		if tokens == nil {
			fmt.Println("未登录")
			return nil
		}

		if flagJSON {
			data, _ := json.MarshalIndent(map[string]interface{}{
				"authorized": true,
				"serverUrl":  tokens.ServerURL,
				"clientId":   tokens.ClientID,
				"scope":      tokens.Scope,
				"expiresAt":  tokens.ExpiresAt.Format(time.RFC3339),
				"expired":    tokens.IsExpired(),
			}, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		color.Green("已登录")
		fmt.Printf("  Server:    %s\n", tokens.ServerURL)
		fmt.Printf("  Client ID: %s\n", tokens.ClientID)
		fmt.Printf("  Scope:     %s\n", tokens.Scope)
		fmt.Printf("  Expires:   %s\n", tokens.ExpiresAt.Format(time.RFC3339))
		if tokens.IsExpired() {
			color.Yellow("  (Token 已过期, 将在下次请求时自动刷新)")
		}
		return nil
	},
}

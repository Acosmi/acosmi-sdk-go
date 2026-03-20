package main

import (
	"fmt"
	"os"

	acosmi "github.com/acosmi/acosmi-sdk-go"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

var (
	flagServer string
	flagJSON   bool
	cliCfg     CLIConfig
	client     *acosmi.Client
)

// needsClient 判断命令是否需要初始化 SDK client
func needsClient(cmd *cobra.Command) bool {
	name := cmd.Name()
	// version/config 等纯本地命令不需要 client
	if name == "version" || name == "completion" {
		return false
	}
	// config 子命令也不需要
	if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
		return false
	}
	if name == "config" {
		return false
	}
	return true
}

var rootCmd = &cobra.Command{
	Use:   "crabclaw",
	Short: "Acosmi 技能商店 CLI",
	Long:  "CrabClaw — Acosmi 技能商店命令行工具\n搜索、安装、上传和管理技能",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !needsClient(cmd) {
			return nil
		}

		cliCfg = loadConfig()

		// --server flag 覆盖
		if flagServer != "" {
			cliCfg.ServerURL = flagServer
		}

		// 创建 SDK client (合并后统一 token: ~/.acosmi/tokens.json)
		c, err := acosmi.NewClient(acosmi.Config{
			ServerURL: cliCfg.ServerURL,
			// Store: nil → 默认 NewFileTokenStore("") → ~/.acosmi/tokens.json
		})
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}
		client = c
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "", "服务器地址 (覆盖配置)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "JSON 输出 (方便脚本集成)")
}

// requireAuth 检查是否已登录, 未登录返回错误
func requireAuth() error {
	if client == nil || !client.IsAuthorized() {
		return fmt.Errorf("未登录。请先运行: crabclaw login")
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

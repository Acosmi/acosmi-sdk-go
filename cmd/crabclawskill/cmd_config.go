package main

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd, configSetCmd, configResetCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理 CLI 配置",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		if flagJSON {
			data, _ := json.MarshalIndent(cfg, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Printf("Server URL: %s\n", cfg.ServerURL)
		fmt.Printf("Skill Dir:  %s\n", cfg.SkillDir)
		fmt.Printf("Config:     %s\n", configPath())
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "设置配置项 (server, skilldir)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfig()

		switch args[0] {
		case "server":
			cfg.ServerURL = args[1]
		case "skilldir":
			cfg.SkillDir = args[1]
		default:
			return fmt.Errorf("未知配置项: %s (可用: server, skilldir)", args[0])
		}

		if err := saveConfig(cfg); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}

		color.Green("已设置 %s = %s", args[0], args[1])
		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "重置为默认配置",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := saveConfig(defaultConfig()); err != nil {
			return fmt.Errorf("重置配置失败: %w", err)
		}
		color.Green("配置已重置为默认值")
		return nil
	},
}

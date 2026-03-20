package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出已安装技能 (需登录)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		ctx := context.Background()

		// 服务端统计
		summary, err := client.GetSkillSummary(ctx)
		if err != nil {
			return fmt.Errorf("获取统计失败: %w", err)
		}

		// 本地已安装
		type localSkill struct {
			Key     string `json:"key"`
			Name    string `json:"name,omitempty"`
			Version string `json:"version,omitempty"`
		}
		var locals []localSkill

		skillDir := cliCfg.SkillDir
		entries, _ := os.ReadDir(skillDir)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			ls := localSkill{Key: entry.Name()}

			// 尝试读 manifest.json
			manifestPath := filepath.Join(skillDir, entry.Name(), "manifest.json")
			if data, err := os.ReadFile(manifestPath); err == nil {
				var m struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}
				if json.Unmarshal(data, &m) == nil {
					ls.Name = m.Name
					ls.Version = m.Version
				}
			}
			locals = append(locals, ls)
		}

		if flagJSON {
			data, _ := json.MarshalIndent(map[string]interface{}{
				"summary": summary,
				"local":   locals,
			}, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("服务端统计: 已安装 %d | 已创建 %d | 总计 %d | 商店可用 %d\n\n",
			summary.Installed, summary.Created, summary.Total, summary.StoreAvailable)

		if len(locals) == 0 {
			fmt.Println("本地无已安装技能")
			return nil
		}

		fmt.Printf("本地已安装 (%s):\n\n", skillDir)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KEY\tNAME\tVERSION")
		fmt.Fprintln(w, "---\t----\t-------")
		for _, ls := range locals {
			name := ls.Name
			if name == "" {
				name = "-"
			}
			ver := ls.Version
			if ver == "" {
				ver = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", ls.Key, name, ver)
		}
		w.Flush()

		return nil
	},
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	searchCmd.Flags().String("category", "", "按分类过滤 (ACTION|TRIGGER|TRANSFORM)")
	searchCmd.Flags().String("tag", "", "按标签过滤")
	searchCmd.Flags().String("source", "", "按来源过滤 (BUILTIN|COMMUNITY|USER)")
	searchCmd.Flags().Int("page", 1, "页码")
	searchCmd.Flags().Int("page-size", 20, "每页数量")
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "搜索公共技能 (无需登录)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword := args[0]
		category, _ := cmd.Flags().GetString("category")
		tag, _ := cmd.Flags().GetString("tag")
		source, _ := cmd.Flags().GetString("source")
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		result, err := client.BrowseSkills(context.Background(), page, pageSize, category, keyword, tag, source)
		if err != nil {
			return fmt.Errorf("搜索失败: %w", err)
		}

		if flagJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(result.Items) == 0 {
			fmt.Println("未找到匹配的技能")
			return nil
		}

		fmt.Printf("找到 %d 个技能 (第 %d 页)\n\n", result.Total, result.Page)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KEY\tNAME\tCATEGORY\tSOURCE\tDOWNLOADS\tSECURITY")
		fmt.Fprintln(w, "---\t----\t--------\t------\t---------\t--------")
		for _, item := range result.Items {
			src := item.Source
			if src == "" {
				src = "COMMUNITY"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
				item.Key, item.Name, item.Category, src, item.DownloadCount, item.SecurityLevel)
		}
		w.Flush()

		return nil
	},
}

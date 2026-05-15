package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "查看配置信息",
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "显示配置文件路径",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(configPath)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configPathCmd)
	rootCmd.AddCommand(configCmd)
}

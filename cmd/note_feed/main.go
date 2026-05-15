package main

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	rootCmd    = &cobra.Command{
		Use:   "note_feed",
		Short: "笔记推送系统 — 基于 Obsidian Vault git commit 的学习资料推送工具",
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "./config.yaml", "配置文件路径")
}

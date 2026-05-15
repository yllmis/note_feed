package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"note_feed/internal/config"
	"note_feed/internal/db"
)

const maxDiffSize = 20 * 1024

var cacheDiffCmd = &cobra.Command{
	Use:   "cache-diff",
	Short: "缓存 diff 内容到数据库（由 post-commit hook 调用）",
	RunE: func(cmd *cobra.Command, args []string) error {
		commitHash, _ := cmd.Flags().GetString("commit-hash")
		if commitHash == "" {
			return fmt.Errorf("--commit-hash 是必填参数")
		}

		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("读取 diff 内容失败: %w", err)
		}

		content := string(data)
		content = strings.TrimSpace(content)
		if content == "" {
			return nil
		}

		if len(content) > maxDiffSize {
			content = content[:maxDiffSize] + "\n\n[注意：diff 过长已被截断]"
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		vaultPath := cfg.VaultPath
		if vaultPath == "" {
			vaultPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("获取当前目录失败: %w", err)
			}
		}

		dbPath := cfg.DB.Path
		if !filepath.IsAbs(dbPath) {
			dbPath = filepath.Join(vaultPath, dbPath)
		}

		dbConn, err := db.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("打开数据库失败: %w", err)
		}
		defer dbConn.Close()

		if err := db.InsertPendingDiff(dbConn, commitHash, content); err != nil {
			return fmt.Errorf("缓存 diff 失败: %w", err)
		}

		fmt.Printf("已缓存 commit %s 的 diff（%d bytes）\n", commitHash[:8], len(content))
		return nil
	},
}

func init() {
	cacheDiffCmd.Flags().String("commit-hash", "", "Git commit hash")
	rootCmd.AddCommand(cacheDiffCmd)
}

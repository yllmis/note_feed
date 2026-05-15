package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"note_feed/internal/db"
	"note_feed/internal/git"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置、数据库和 git hook",
	RunE: func(cmd *cobra.Command, args []string) error {
		vaultPath, _ := cmd.Flags().GetString("vault")
		if vaultPath == "" {
			var err error
			vaultPath, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("无法获取当前目录: %w", err)
			}
		}
		fmt.Printf("Vault 路径: %s\n", vaultPath)

		if !git.IsGitRepo(vaultPath) {
			return fmt.Errorf("%s 不是 git 仓库，请先 git init", vaultPath)
		}
		fmt.Println("✓ git 仓库检测通过")

		// 写入 hook（嵌入 config 绝对路径，让 cache-diff 能找到配置）
		cfgAbsPath, err := filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("获取配置绝对路径失败: %w", err)
		}

		hookPath := filepath.Join(vaultPath, ".git", "hooks", "post-commit")
		hookContent := fmt.Sprintf(`#!/bin/bash
# 笔记推送系统 - post-commit hook
git diff HEAD~1 HEAD -- '*.md' | note_feed --config %s cache-diff --commit-hash=$(git rev-parse HEAD)
`, cfgAbsPath)
		if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
			return fmt.Errorf("写入 post-commit hook 失败: %w", err)
		}
		fmt.Printf("✓ post-commit hook 已安装（config: %s）\n", cfgAbsPath)

		// 初始化数据库
		dbPath := filepath.Join(vaultPath, "data", "push.db")
		if err := os.MkdirAll(filepath.Join(vaultPath, "data"), 0755); err != nil {
			return fmt.Errorf("创建 data 目录失败: %w", err)
		}
		dbConn, err := db.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("初始化数据库失败: %w", err)
		}
		defer dbConn.Close()
		fmt.Println("✓ 数据库已初始化:", dbPath)

		// 创建配置文件（到当前目录，即项目目录）
		cfgContent := fmt.Sprintf(`# 笔记推送系统配置
vault_path: "%s"

llm:
  api_key: "${DEEPSEEK_API_KEY}"
  model: "deepseek-chat"
  timeout: "30s"
  base_url: "https://api.deepseek.com"

search:
  juejin:
    enabled: true
  google:
    enabled: false
    api_key: "${GOOGLE_API_KEY}"
    cse_id: "${GOOGLE_CSE_ID}"

push:
  email:
    smtp_host: "smtp.qq.com"
    smtp_port: 587
    username: "${EMAIL_USER}"
    password: "${EMAIL_PASS}"
    from: "${EMAIL_USER}"
    to: "${EMAIL_TO}"

db:
  path: "%s"
`, vaultPath, dbPath)

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := os.WriteFile(configPath, []byte(cfgContent), 0644); err != nil {
				return fmt.Errorf("写入配置文件失败: %w", err)
			}
			fmt.Println("✓ 配置文件已创建:", cfgAbsPath)
		} else {
			fmt.Println("✓ 配置文件已存在，跳过:", cfgAbsPath)
		}

		fmt.Println("\n✅ 初始化完成！请设置环境变量:")
		fmt.Println("   export DEEPSEEK_API_KEY=sk-xxx")
		fmt.Println("   export EMAIL_USER=xxx@qq.com")
		fmt.Println("   export EMAIL_PASS=授权码")
		fmt.Println("   export EMAIL_TO=xxx@qq.com")
		return nil
	},
}

func init() {
	initCmd.Flags().StringP("vault", "v", "", "Vault 路径（默认当前目录）")
	rootCmd.AddCommand(initCmd)
}

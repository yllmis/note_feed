package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"note_feed/internal/config"
	"note_feed/internal/db"
	"note_feed/internal/git"
	"note_feed/internal/llm"
	"note_feed/internal/push"
	"note_feed/internal/search"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "执行推送（每日学习资料推送）",
}

var pushDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "聚合当日学习内容并推送",
	RunE: func(cmd *cobra.Command, args []string) error {
		dateStr, _ := cmd.Flags().GetString("date")
		if dateStr == "" {
			dateStr = time.Now().Format("2006-01-02")
		}

		vaultOverride, _ := cmd.Flags().GetString("vault")

		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		vaultPath := vaultOverride
		if vaultPath == "" {
			vaultPath = cfg.VaultPath
		}
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

		diffs, err := db.GetPendingDiffsByDate(dbConn, dateStr)
		if err != nil {
			return fmt.Errorf("查询当日 diffs 失败: %w", err)
		}

		if len(diffs) == 0 {
			fmt.Println("pending_diffs 为空，尝试从 git log 拉取...")
			hashes, err := git.GetTodayCommits(vaultPath, dateStr)
			if err != nil {
				return fmt.Errorf("获取今日 git 提交失败: %w", err)
			}
			for _, h := range hashes {
				diffContent, err := git.DiffForCommit(vaultPath, h)
				if err != nil {
					continue
				}
				_ = db.InsertPendingDiff(dbConn, h, diffContent)
				diffs = append(diffs, db.PendingDiff{
					CommitHash:  h,
					DiffContent: diffContent,
				})
			}
		}

		if len(diffs) == 0 {
			fmt.Println("今日无学习记录，跳过推送")
			return nil
		}

		var allDiffs []string
		var diffIDs []int64
		for _, d := range diffs {
			allDiffs = append(allDiffs, d.DiffContent)
			diffIDs = append(diffIDs, d.ID)
		}
		mergedDiff := strings.Join(allDiffs, "\n\n---\n\n")

		timeout, _ := time.ParseDuration(cfg.LLM.Timeout)
		llmClient := llm.NewClient(cfg.LLM.APIKey, cfg.LLM.Model, cfg.LLM.BaseURL, timeout)

		fmt.Println("正在提炼知识点...")
		extraction, err := llmClient.ExtractTopics(mergedDiff)
		if err != nil {
			return fmt.Errorf("提炼知识点失败: %w", err)
		}

		if len(extraction.Topics) == 0 {
			fmt.Println("未提炼出知识点，跳过推送")
			return nil
		}

		fmt.Println("正在搜索相关文章...")
		var allResults []*search.SearchResult
		for _, topic := range extraction.Topics {
			result, err := search.SearchByCategory(dbConn, llmClient, topic, 3)
			if err != nil {
				fmt.Printf("搜索 [%s] 失败: %v\n", topic.Category, err)
				continue
			}
			if len(result.Articles) > 0 {
				allResults = append(allResults, result)
			}
		}

		if len(allResults) == 0 {
			fmt.Println("未搜索到有效文章，跳过推送")
			return nil
		}

		allArticles := search.DedupArticles(allResults)
		fmt.Println("正在生成文章摘要...")
		_ = search.GenerateSummaries(llmClient, allArticles)

		var categories []string
		for _, t := range extraction.Topics {
			categories = append(categories, t.Category)
		}

		fmt.Println("正在推送...")
		if err := push.SendDigest(*cfg, allArticles, categories, dateStr); err != nil {
			_ = db.InsertPushLog(dbConn, dateStr, getTopicSummary(extraction), len(allArticles), "failed")
			return fmt.Errorf("推送失败: %w", err)
		}

		for _, a := range allArticles {
			_ = db.InsertPushedArticle(dbConn, a.URL, a.Title, a.Category)
		}
		_ = db.DeletePendingDiffs(dbConn, diffIDs)
		_ = db.InsertPushLog(dbConn, dateStr, getTopicSummary(extraction), len(allArticles), "success")

		fmt.Printf("推送完成！共 %d 篇文章，%d 个分类\n", len(allArticles), len(categories))
		return nil
	},
}

var pushTestCmd = &cobra.Command{
	Use:   "test",
	Short: "发送测试邮件验证配置",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		if cfg.Push.Email.Username == "" || cfg.Push.Email.Password == "" {
			return fmt.Errorf("请设置邮箱账号和密码（环境变量 EMAIL_USER / EMAIL_PASS）")
		}

		fmt.Println("正在发送测试邮件...")
		if err := push.SendTestEmail(cfg.Push.Email); err != nil {
			return fmt.Errorf("发送测试邮件失败: %w", err)
		}
		fmt.Println("测试邮件已发送，请查收！")
		return nil
	},
}

func init() {
	pushDailyCmd.Flags().String("date", "", "指定日期 (YYYY-MM-DD)，默认今日")
	pushDailyCmd.Flags().StringP("vault", "v", "", "覆盖 Vault 路径（默认从配置读取）")
	pushTestCmd.Flags().StringP("vault", "v", "", "覆盖 Vault 路径（默认从配置读取）")
	pushCmd.AddCommand(pushDailyCmd)
	pushCmd.AddCommand(pushTestCmd)
	rootCmd.AddCommand(pushCmd)
}

func getTopicSummary(extraction *llm.TopicExtraction) string {
	var parts []string
	for _, t := range extraction.Topics {
		parts = append(parts, fmt.Sprintf("%s: %s", t.Category, strings.Join(t.Keywords, ", ")))
	}
	return strings.Join(parts, "; ")
}

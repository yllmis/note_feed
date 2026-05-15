package search

import (
	"database/sql"
	"fmt"
	"strings"

	"note_feed/internal/db"
	"note_feed/internal/llm"
)

var categoryCN = map[string]string{
	"10-Computer-Science": "计算机基础",
	"20-Go-Ecosystem":     "Go语言",
	"30-Backend-Infra":    "后端中间件",
	"40-Architecture":     "系统架构",
	"50-Projects":         "项目实战",
	"60-AI-Agent":         "AI智能体",
	"70-LeetCode":         "算法刷题",
}

func CategoryChinese(category string) string {
	if cn, ok := categoryCN[category]; ok {
		return cn
	}
	return category
}

type SearchResult struct {
	Category string
	Articles []Article
}

func SearchByCategory(
	dbConn *sql.DB,
	llmClient *llm.Client,
	topic llm.Topic,
	maxArticles int,
	tavilyAPIKey string,
) (*SearchResult, error) {
	if tavilyAPIKey == "" {
		return nil, fmt.Errorf("Tavily API Key 未配置")
	}

	categoryCNStr := CategoryChinese(topic.Category)

	var keywords []string
	if len(topic.Keywords) > 2 {
		keywords = topic.Keywords[:2]
	} else {
		keywords = topic.Keywords
	}

	query := categoryCNStr + " " + strings.Join(keywords, " ")

	articles, err := SearchTavily(tavilyAPIKey, query, maxArticles*2)
	if err != nil {
		return nil, fmt.Errorf("搜索 [%s] 失败: %w", topic.Category, err)
	}

	var filtered []Article
	for _, a := range articles {
		if len(filtered) >= maxArticles {
			break
		}
		pushed, err := db.IsArticlePushed(dbConn, a.URL)
		if err != nil || pushed {
			continue
		}
		a.Category = topic.Category
		filtered = append(filtered, a)
	}

	return &SearchResult{
		Category: topic.Category,
		Articles: filtered,
	}, nil
}

func GenerateSummaries(llmClient *llm.Client, articles []Article) error {
	for i := range articles {
		if articles[i].Description == "" {
			continue
		}
		summary, err := llmClient.GenerateSummary(articles[i].Title, articles[i].Description)
		if err != nil {
			continue
		}
		articles[i].Description = summary
	}
	return nil
}

func DedupArticles(results []*SearchResult) []Article {
	seen := make(map[string]bool)
	var all []Article
	for _, r := range results {
		for _, a := range r.Articles {
			if seen[a.URL] {
				continue
			}
			seen[a.URL] = true
			all = append(all, a)
		}
	}
	return all
}

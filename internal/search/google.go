package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type googleResponse struct {
	Items []googleItem `json:"items"`
}

type googleItem struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

func SearchGoogle(apiKey, cseID, query string, limit int) ([]Article, error) {
	if apiKey == "" || cseID == "" {
		return nil, fmt.Errorf("Google API Key 或 CSE ID 未配置")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	params := url.Values{}
	params.Set("key", apiKey)
	params.Set("cx", cseID)
	params.Set("q", query)
	params.Set("num", fmt.Sprintf("%d", limit))
	params.Set("lr", "lang_zh-CN")
	params.Set("hl", "zh-CN")

	apiURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?%s", params.Encode())

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Google 搜索请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 Google 搜索 API 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 Google 响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Google API 返回错误状态 %d: %s", resp.StatusCode, string(body))
	}

	var result googleResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 Google 响应失败: %w", err)
	}

	var articles []Article
	for _, item := range result.Items {
		title := strings.TrimSpace(item.Title)
		snippet := strings.TrimSpace(item.Snippet)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}

		articles = append(articles, Article{
			Title:       title,
			URL:         item.Link,
			Description: snippet,
		})
	}

	return articles, nil
}

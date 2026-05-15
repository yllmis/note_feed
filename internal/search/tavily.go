package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type tavilyRequest struct {
	APIKey   string `json:"api_key"`
	Query    string `json:"query"`
	MaxCount int    `json:"max_results"`
	Topic    string `json:"topic,omitempty"`
}

type tavilyResponse struct {
	Results []tavilyResult `json:"results"`
}

type tavilyResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

func SearchTavily(apiKey, query string, limit int) ([]Article, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Tavily API Key 未配置")
	}

	client := &http.Client{Timeout: 15 * time.Second}

	body := tavilyRequest{
		APIKey:   apiKey,
		Query:    query,
		MaxCount: limit,
		Topic:    "general",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.tavily.com/search", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 Tavily API 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Tavily API 返回错误状态 %d: %s", resp.StatusCode, string(respBody))
	}

	var result tavilyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	var articles []Article
	for _, r := range result.Results {
		content := r.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		articles = append(articles, Article{
			Title:       r.Title,
			URL:         r.URL,
			Description: content,
		})
	}

	return articles, nil
}

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

type Article struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type juejinResponse struct {
	Data []juejinItem `json:"data"`
}

type juejinItem struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Content     string `json:"content"`
	Description string `json:"description"`
}

func SearchJuejin(keyword string, limit int) ([]Article, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	params := url.Values{}
	params.Set("keyWord", keyword)
	params.Set("page", "0")
	params.Set("pageSize", fmt.Sprintf("%d", limit))

	apiURL := fmt.Sprintf("https://api.juejin.cn/search_api/v1/search?%s", params.Encode())

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建掘金搜索请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用掘金搜索 API 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取掘金响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("掘金 API 返回错误状态 %d", resp.StatusCode)
	}

	var result juejinResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析掘金响应失败: %w", err)
	}

	var articles []Article
	for _, item := range result.Data {
		title := cleanHTMLTags(item.Title)
		desc := item.Description
		if desc == "" {
			desc = cleanHTMLTags(item.Content)
		}
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}

		articles = append(articles, Article{
			Title:       title,
			URL:         item.URL,
			Description: desc,
		})
	}

	if len(articles) > limit {
		articles = articles[:limit]
	}

	return articles, nil
}

func cleanHTMLTags(s string) string {
	s = strings.ReplaceAll(s, "<em>", "")
	s = strings.ReplaceAll(s, "</em>", "")
	return s
}

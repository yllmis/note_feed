package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	apiKey  string
	model   string
	baseURL string
	timeout time.Duration
	client  *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type TopicExtraction struct {
	Topics []Topic `json:"topics"`
}

type Topic struct {
	Category   string   `json:"category"`
	Keywords   []string `json:"keywords"`
	Extensions []string `json:"extensions"`
}

func NewClient(apiKey, model, baseURL string, timeout time.Duration) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

const extractPrompt = `你是一个知识提炼助手。请分析以下 git diff 内容，提取其中新增或修改的知识点。

要求：
1. 每个知识点归到对应的大类（类别从 vault 文件夹层级来，例如：20-Go-Ecosystem、30-Backend-Infra、40-Architecture 等）
2. 尽量提取具体的技术概念名作为关键词
3. 每个大类列举 1-2 个可延伸拓展的方向
4. 只输出 JSON，不要多余文字

输出格式：
{
  "topics": [
    {
      "category": "30-Backend-Infra",
      "keywords": ["Redis雪崩", "Redis击穿"],
      "extensions": ["Redis集群方案", "热点Key探测"]
    }
  ]
}`

func (c *Client) ExtractTopics(diffText string) (*TopicExtraction, error) {
	if len(diffText) > 20000 {
		diffText = diffText[:20000] + "\n\n[注意：diff 过长已被截断]"
	}

	messages := []Message{
		{Role: "system", Content: extractPrompt},
		{Role: "user", Content: fmt.Sprintf("以下是要分析的 diff 内容：\n\n```\n%s\n```", diffText)},
	}

	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.1,
	}

	var result ChatResponse
	if err := c.callAPI(req, &result); err != nil {
		return nil, fmt.Errorf("LLM 提炼知识点失败: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("LLM 返回空结果")
	}

	content := result.Choices[0].Message.Content
	content = cleanResponse(content)

	var extraction TopicExtraction
	if err := json.Unmarshal([]byte(content), &extraction); err != nil {
		return nil, fmt.Errorf("解析 LLM 输出 JSON 失败: %w\n原始内容: %s", err, content)
	}

	return &extraction, nil
}

const summaryPrompt = `你是一个文章摘要助手。请为以下技术文章生成 1-2 句中文摘要，简洁概括文章核心内容，不要超过 100 字。

文章标题：%s
文章简介：%s

摘要：`

func (c *Client) GenerateSummary(title, description string) (string, error) {
	messages := []Message{
		{Role: "user", Content: fmt.Sprintf(summaryPrompt, title, description)},
	}

	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.3,
	}

	var result ChatResponse
	if err := c.callAPI(req, &result); err != nil {
		return "", fmt.Errorf("LLM 生成摘要失败: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("LLM 返回空结果")
	}

	summary := result.Choices[0].Message.Content
	summary = cleanResponse(summary)
	if len(summary) > 150 {
		summary = summary[:150] + "..."
	}

	return summary, nil
}

func (c *Client) callAPI(req ChatRequest, result *ChatResponse) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("调用 LLM API 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取 LLM 响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("LLM API 返回错误状态 %d: %s", resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("解析 LLM 响应 JSON 失败: %w", err)
	}

	return nil
}

func cleanResponse(s string) string {
	b := []byte(s)
	b = bytes.TrimSpace(b)
	b = bytes.TrimPrefix(b, []byte("```json"))
	b = bytes.TrimPrefix(b, []byte("```"))
	b = bytes.TrimSuffix(b, []byte("```"))
	b = bytes.TrimSpace(b)
	return string(b)
}

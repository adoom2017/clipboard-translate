package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ClaudeClient Claude客户端
type ClaudeClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// ClaudeRequest Claude API请求结构
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	System    string          `json:"system,omitempty"`
}

// ClaudeMessage Claude消息结构
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse Claude API响应结构
type ClaudeResponse struct {
	Content []ClaudeContent `json:"content"`
	Error   *APIError       `json:"error,omitempty"`
}

// ClaudeContent Claude内容结构
type ClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewClaudeClient 创建Claude客户端
func NewClaudeClient(config AIConfig) (*ClaudeClient, error) {
	if config.APIKey == "" {
		return nil, ErrInvalidAPIKey
	}

	if config.Model == "" {
		config.Model = "claude-3-haiku-20240307"
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	return &ClaudeClient{
		apiKey:  config.APIKey,
		model:   config.Model,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Translate 实现翻译功能
func (c *ClaudeClient) Translate(ctx context.Context, text string, isChinese bool) (string, error) {
	request := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 1000,
		System:    getSystemPrompt(isChinese),
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: text,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if claudeResp.Error != nil {
		return "", fmt.Errorf("Claude API错误: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("没有返回翻译结果")
	}

	return claudeResp.Content[0].Text, nil
}

// GetName 获取客户端名称
func (c *ClaudeClient) GetName() string {
	return "Claude"
}

// Close 关闭客户端
func (c *ClaudeClient) Close() error {
	return nil
}

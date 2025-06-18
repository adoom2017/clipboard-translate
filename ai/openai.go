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

// OpenAIClient OpenAI客户端
type OpenAIClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// OpenAIRequest OpenAI API请求结构
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse OpenAI API响应结构
type OpenAIResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice 选择结构
type Choice struct {
	Message Message `json:"message"`
}

// APIError API错误结构
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// NewOpenAIClient 创建OpenAI客户端
func NewOpenAIClient(config AIConfig) (*OpenAIClient, error) {
	if config.APIKey == "" {
		return nil, ErrInvalidAPIKey
	}

	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIClient{
		apiKey:  config.APIKey,
		model:   config.Model,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Translate 实现翻译功能
func (o *OpenAIClient) Translate(ctx context.Context, text string) (string, error) {
	request := OpenAIRequest{
		Model: o.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: getSystemPrompt(),
			},
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

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("OpenAI API错误: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("没有返回翻译结果")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// GetName 获取客户端名称
func (o *OpenAIClient) GetName() string {
	return "OpenAI"
}

// Close 关闭客户端
func (o *OpenAIClient) Close() error {
	// HTTP客户端不需要显式关闭
	return nil
}

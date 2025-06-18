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

// OllamaClient Ollama客户端
type OllamaClient struct {
	model   string
	baseURL string
	client  *http.Client
}

// OllamaRequest Ollama API请求结构
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse Ollama API响应结构
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewOllamaClient 创建Ollama客户端
func NewOllamaClient(config AIConfig) (*OllamaClient, error) {
	if config.Model == "" {
		config.Model = "llama2"
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaClient{
		model:   config.Model,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// Translate 实现翻译功能
func (o *OllamaClient) Translate(ctx context.Context, text string) (string, error) {
	prompt := fmt.Sprintf("%s\n\n%s", getSystemPrompt(), text)

	request := OllamaRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	return ollamaResp.Response, nil
}

// GetName 获取客户端名称
func (o *OllamaClient) GetName() string {
	return "Ollama"
}

// Close 关闭客户端
func (o *OllamaClient) Close() error {
	return nil
}

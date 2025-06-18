package ai

import (
	"context"
	"fmt"
	"time"

	gemini "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient Gemini AI客户端
type GeminiClient struct {
	client *gemini.Client
	model  string
}

// NewGeminiClient 创建Gemini客户端
func NewGeminiClient(config AIConfig) (*GeminiClient, error) {
	if config.APIKey == "" {
		return nil, ErrInvalidAPIKey
	}

	if config.Model == "" {
		config.Model = "models/gemini-2.0-flash"
	} else {
		config.Model = "models/" + config.Model
	}

	ctx := context.Background()
	client, err := gemini.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("创建Gemini客户端失败: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  config.Model,
	}, nil
}

// Translate 实现翻译功能
func (g *GeminiClient) Translate(ctx context.Context, text string, isChinese bool) (string, error) {
	model := g.client.GenerativeModel(g.model)

	systemInstruction := &gemini.Content{
		Parts: []gemini.Part{
			gemini.Text(getSystemPrompt(isChinese)),
		},
		Role: "system",
	}

	model.SystemInstruction = systemInstruction

	// 最大重试次数
	maxRetries := 3
	var lastErr error

	for attempt := range maxRetries {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resp, err := model.GenerateContent(ctx, gemini.Text(text))
		if err != nil {
			lastErr = err
			continue
		}

		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			if responseText, ok := resp.Candidates[0].Content.Parts[0].(gemini.Text); ok {
				return string(responseText), nil
			}
		}
	}

	return "", fmt.Errorf("翻译失败，已重试%d次: %v", maxRetries, lastErr)
}

// GetName 获取客户端名称
func (g *GeminiClient) GetName() string {
	return "Gemini"
}

// Close 关闭客户端
func (g *GeminiClient) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

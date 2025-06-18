package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// AIClient 定义AI客户端接口
type AIClient interface {
	// Translate 翻译文本
	Translate(ctx context.Context, text string, isChinese bool) (string, error)
	// GetName 获取客户端名称
	GetName() string
	// Close 关闭客户端连接
	Close() error
}

// AIConfig AI配置
type AIConfig struct {
	Provider string `json:"provider"` // gemini, openai, claude, ollama, etc.
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url,omitempty"` // 可选的自定义API端点
}

// NewAIClient 创建AI客户端工厂函数
func NewAIClient(config AIConfig) (AIClient, error) {
	switch strings.ToLower(config.Provider) {
	case "gemini":
		return NewGeminiClient(config)
	case "openai":
		return NewOpenAIClient(config)
	case "claude":
		return NewClaudeClient(config)
	case "ollama":
		return NewOllamaClient(config)
	default:
		return nil, fmt.Errorf("不支持的AI提供商: %s", config.Provider)
	}
}

// 获取系统提示词
func getSystemPrompt(isChinese bool) string {
	if isChinese {
		return "你是一个专业的中英文翻译助手，请将用户输入的中文内容翻译成英文，只需给出翻译结果，不要输出多余内容。"
	} else {
		return "你是一个专业的英中文翻译助手，请将用户输入的英文内容翻译成中文，只需给出翻译结果，不要输出多余内容。"
	}
}

// 通用错误定义
var (
	ErrInvalidAPIKey     = errors.New("无效的API密钥")
	ErrNetworkError      = errors.New("网络连接错误")
	ErrRateLimitExceeded = errors.New("API调用频率限制")
	ErrModelNotFound     = errors.New("模型不存在")
)

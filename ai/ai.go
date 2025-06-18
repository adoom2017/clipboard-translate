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
	Translate(ctx context.Context, text string) (string, error)
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
func getSystemPrompt() string {
	return `角色设定与目标：
* 你是一位专业的翻译专家，精通中文和英文之间的互译。
* 你的主要目标是准确、流畅地完成用户提供的中英文翻译任务。
* 你能自动识别用户输入的语言类型（中文或英文）。
* 根据识别结果，将输入的文本翻译成对应的目标语言（中文输入则翻译成英文，英文输入则翻译成中文）。
* 你的回复只包含翻译结果，不包含任何额外的解释或说明。

行为准则：
1)  语言识别：
    * 准确判断用户输入的文本是中文还是英文。
    * 如果输入包含中英文混合文本，你需要判断主要语种并将其翻译为另一种语言。

2)  翻译执行：
    * 使用清晰、准确、自然的语言进行翻译。
    * 力求在翻译过程中保留原文的含义和语境。
    * 对于口语化的表达，翻译应贴近日常用法。
    * 对于专业术语，应使用行业内通用的翻译。

3)  输出格式：
    * 只输出翻译后的文本内容。
    * 避免在翻译结果前后添加任何不必要的文字、符号或提示。

沟通方式：
* 简洁明了，直接给出翻译结果。`
}

// 通用错误定义
var (
	ErrInvalidAPIKey     = errors.New("无效的API密钥")
	ErrNetworkError      = errors.New("网络连接错误")
	ErrRateLimitExceeded = errors.New("API调用频率限制")
	ErrModelNotFound     = errors.New("模型不存在")
)

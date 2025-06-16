package database

import (
	"fmt"
	"time"
)

// HistoryItem 代表翻译历史中的一项记录
type HistoryItem struct {
	ID         string    `json:"id"`
	Original   string    `json:"original"`
	Translated string    `json:"translated"`
	Direction  string    `json:"direction"` // 翻译方向，如 "中 → 英"
	Timestamp  time.Time `json:"timestamp"`
}

// Database 定义数据库操作的接口
type Database interface {
	// 初始化数据库连接和表结构
	Initialize() error

	// 关闭数据库连接
	Close() error

	// 添加新的翻译历史记录
	AddHistoryItem(item *HistoryItem) error

	// 获取所有历史记录，按时间倒序排列
	GetHistoryItems() ([]*HistoryItem, error)

	// 清空所有历史记录
	ClearHistory() error

	// 获取历史记录数量
	GetHistoryCount() (int, error)

	// 删除超出保留数量的旧记录
	PruneHistory(keepCount int) error
}

// 数据库配置结构
type DBConfig struct {
	Type       string `json:"type"`
	Connection string `json:"connection"`
	MaxHistory int    `json:"max_history"`
}

// New 创建指定类型的数据库实现
func New(config DBConfig) (Database, error) {
	switch config.Type {
	case "sqlite", "":
		return NewSQLiteDB(config.Connection)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}
}

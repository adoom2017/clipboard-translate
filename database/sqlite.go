package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB SQLite数据库实现
type SQLiteDB struct {
	db     *sql.DB
	dbPath string
}

// NewSQLiteDB 创建一个新的SQLite数据库连接
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	if dbPath == "" {
		dbPath = "clipboard-translate.db"
	}

	return &SQLiteDB{dbPath: dbPath}, nil
}

// Initialize 初始化数据库连接和表结构
func (s *SQLiteDB) Initialize() error {
	var err error

	// 打开数据库连接
	s.db, err = sql.Open("sqlite3", s.dbPath+"?_journal=WAL&_timeout=5000&_fk=true")
	if err != nil {
		return fmt.Errorf("无法打开SQLite数据库: %w", err)
	}

	// 设置连接池参数
	s.db.SetMaxOpenConns(1) // SQLite只支持单连接
	s.db.SetMaxIdleConns(1)

	// 创建表结构
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id TEXT PRIMARY KEY,
			original TEXT NOT NULL,
			translated TEXT NOT NULL,
			direction TEXT NOT NULL,
			timestamp DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_history_timestamp ON history(timestamp DESC);
	`)

	if err != nil {
		return fmt.Errorf("创建表结构失败: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (s *SQLiteDB) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// AddHistoryItem 添加新的翻译历史记录
func (s *SQLiteDB) AddHistoryItem(item *HistoryItem) error {
	_, err := s.db.Exec(
		"INSERT INTO history (id, original, translated, direction, timestamp) VALUES (?, ?, ?, ?, ?)",
		item.ID,
		item.Original,
		item.Translated,
		item.Direction,
		item.Timestamp,
	)

	return err
}

// GetHistoryItems 获取所有历史记录，按时间倒序排列
func (s *SQLiteDB) GetHistoryItems() ([]*HistoryItem, error) {
	rows, err := s.db.Query("SELECT id, original, translated, direction, timestamp FROM history ORDER BY timestamp DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*HistoryItem

	for rows.Next() {
		item := &HistoryItem{}
		var timestamp string

		err := rows.Scan(&item.ID, &item.Original, &item.Translated, &item.Direction, &timestamp)
		if err != nil {
			return nil, err
		}

		item.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestamp)
		if err != nil {
			// 如果解析失败，使用当前时间
			item.Timestamp = time.Now()
		}

		items = append(items, item)
	}

	return items, nil
}

// ClearHistory 清空所有历史记录
func (s *SQLiteDB) ClearHistory() error {
	_, err := s.db.Exec("DELETE FROM history")
	return err
}

// GetHistoryCount 获取历史记录数量
func (s *SQLiteDB) GetHistoryCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM history").Scan(&count)
	return count, err
}

// PruneHistory 删除超出保留数量的旧记录
func (s *SQLiteDB) PruneHistory(keepCount int) error {
	if keepCount <= 0 {
		return nil
	}

	_, err := s.db.Exec(`
		DELETE FROM history
		WHERE id IN (
			SELECT id FROM history
			ORDER BY timestamp DESC
			LIMIT -1 OFFSET ?
		)
	`, keepCount)

	return err
}

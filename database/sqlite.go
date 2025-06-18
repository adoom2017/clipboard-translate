package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
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

	s.db, err = sql.Open("sqlite", s.dbPath)
	if err != nil {
		return fmt.Errorf("无法打开SQLite数据库: %w", err)
	}

	// 设置连接池参数
	s.db.SetMaxOpenConns(1) // SQLite只支持单连接
	s.db.SetMaxIdleConns(1)

	// 创建表结构 - 使用 INTEGER 类型存储时间戳（秒）
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id TEXT PRIMARY KEY,
			original TEXT NOT NULL,
			translated TEXT NOT NULL,
			direction TEXT NOT NULL,
			timestamp INTEGER NOT NULL  -- 存储 Unix 时间戳（秒）
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
	// 将时间转换为 Unix 时间戳（秒）
	unixTimestamp := item.Timestamp.Unix()

	_, err := s.db.Exec(
		"INSERT INTO history (id, original, translated, direction, timestamp) VALUES (?, ?, ?, ?, ?)",
		item.ID,
		item.Original,
		item.Translated,
		item.Direction,
		unixTimestamp, // 存储为秒
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
		var timestamp int64 // 使用 int64 类型读取 Unix 时间戳

		err := rows.Scan(&item.ID, &item.Original, &item.Translated, &item.Direction, &timestamp)
		if err != nil {
			return nil, err
		}

		// 将 Unix 时间戳转换回 time.Time
		item.Timestamp = time.Unix(timestamp, 0)

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

// GetHistoryByDateRange 根据日期范围查询历史记录
func (s *SQLiteDB) GetHistoryByDateRange(start, end time.Time) ([]*HistoryItem, error) {
	// 转换为 Unix 时间戳
	startTS := start.Unix()
	endTS := end.Unix()

	rows, err := s.db.Query(
		"SELECT id, original, translated, direction, timestamp FROM history WHERE timestamp >= ? AND timestamp <= ? ORDER BY timestamp DESC",
		startTS,
		endTS,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*HistoryItem

	for rows.Next() {
		item := &HistoryItem{}
		var timestamp int64 // 使用 int64 类型读取 Unix 时间戳

		err := rows.Scan(&item.ID, &item.Original, &item.Translated, &item.Direction, &timestamp)
		if err != nil {
			return nil, err
		}

		// 将 Unix 时间戳转换回 time.Time
		item.Timestamp = time.Unix(timestamp, 0)

		items = append(items, item)
	}

	return items, nil
}

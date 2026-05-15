package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("获取数据库路径失败: %w", err)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	db, err := sql.Open("sqlite", absPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("初始化数据表失败: %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS pending_diffs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			commit_hash TEXT NOT NULL,
			diff_content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pushed_articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			category TEXT NOT NULL,
			pushed_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS push_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			push_date DATE NOT NULL,
			topic_summary TEXT,
			article_count INTEGER,
			status TEXT DEFAULT 'success',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("执行建表语句失败: %w", err)
		}
	}
	return nil
}

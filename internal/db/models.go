package db

import (
	"database/sql"
	"fmt"
	"time"
)

type PendingDiff struct {
	ID          int64     `json:"id"`
	CommitHash  string    `json:"commit_hash"`
	DiffContent string    `json:"diff_content"`
	CreatedAt   time.Time `json:"created_at"`
}

type PushedArticle struct {
	ID       int64     `json:"id"`
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	Category string    `json:"category"`
	PushedAt time.Time `json:"pushed_at"`
}

type PushLog struct {
	ID           int64     `json:"id"`
	PushDate     string    `json:"push_date"`
	TopicSummary string    `json:"topic_summary"`
	ArticleCount int       `json:"article_count"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func InsertPendingDiff(db *sql.DB, commitHash, diffContent string) error {
	_, err := db.Exec(
		`INSERT INTO pending_diffs (commit_hash, diff_content) VALUES (?, ?)`,
		commitHash, diffContent,
	)
	if err != nil {
		return fmt.Errorf("插入 pending diff 失败: %w", err)
	}
	return nil
}

func GetTodayPendingDiffs(db *sql.DB) ([]PendingDiff, error) {
	rows, err := db.Query(
		`SELECT id, commit_hash, diff_content, created_at
		 FROM pending_diffs
		 WHERE date(created_at) = date('now')
		 ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("查询 today diffs 失败: %w", err)
	}
	defer rows.Close()

	var diffs []PendingDiff
	for rows.Next() {
		var d PendingDiff
		if err := rows.Scan(&d.ID, &d.CommitHash, &d.DiffContent, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描 diff 行失败: %w", err)
		}
		diffs = append(diffs, d)
	}
	return diffs, rows.Err()
}

func GetPendingDiffsByDate(db *sql.DB, date string) ([]PendingDiff, error) {
	rows, err := db.Query(
		`SELECT id, commit_hash, diff_content, created_at
		 FROM pending_diffs
		 WHERE date(created_at) = ?
		 ORDER BY created_at ASC`,
		date,
	)
	if err != nil {
		return nil, fmt.Errorf("查询指定日期 diffs 失败: %w", err)
	}
	defer rows.Close()

	var diffs []PendingDiff
	for rows.Next() {
		var d PendingDiff
		if err := rows.Scan(&d.ID, &d.CommitHash, &d.DiffContent, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描 diff 行失败: %w", err)
		}
		diffs = append(diffs, d)
	}
	return diffs, rows.Err()
}

func DeletePendingDiffs(db *sql.DB, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	// Build placeholders
	query := "DELETE FROM pending_diffs WHERE id IN (?"
	args := []interface{}{ids[0]}
	for i := 1; i < len(ids); i++ {
		query += ", ?"
		args = append(args, ids[i])
	}
	query += ")"

	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("删除 pending diffs 失败: %w", err)
	}
	return nil
}

func IsArticlePushed(db *sql.DB, url string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM pushed_articles WHERE url = ?`, url).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("查询文章去重失败: %w", err)
	}
	return count > 0, nil
}

func InsertPushedArticle(db *sql.DB, url, title, category string) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO pushed_articles (url, title, category) VALUES (?, ?, ?)`,
		url, title, category,
	)
	if err != nil {
		return fmt.Errorf("插入已推送文章记录失败: %w", err)
	}
	return nil
}

func InsertPushLog(db *sql.DB, pushDate, topicSummary string, articleCount int, status string) error {
	_, err := db.Exec(
		`INSERT INTO push_logs (push_date, topic_summary, article_count, status) VALUES (?, ?, ?, ?)`,
		pushDate, topicSummary, articleCount, status,
	)
	if err != nil {
		return fmt.Errorf("插入推送日志失败: %w", err)
	}
	return nil
}

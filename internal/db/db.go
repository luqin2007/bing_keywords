package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

type Keyword struct {
	ID        int64
	Word      string
	Source    string
	CreatedAt int64
	UsedAt    int64
}

func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS keywords (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			word       TEXT    NOT NULL UNIQUE,
			source     TEXT    NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL,
			used_at    INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_keywords_used_at ON keywords(used_at);
		CREATE INDEX IF NOT EXISTS idx_keywords_created_at ON keywords(created_at);
		CREATE TABLE IF NOT EXISTS meta (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}
	_, err = db.conn.Exec("ALTER TABLE keywords ADD COLUMN source TEXT NOT NULL DEFAULT ''")
	if err != nil {
		if !isColumnExistsErr(err) {
			return err
		}
	}
	return nil
}

func isColumnExistsErr(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate column") ||
		contains(err.Error(), "already exists"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (db *DB) Count() (int, error) {
	var n int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM keywords").Scan(&n)
	return n, err
}

func (db *DB) Insert(words []string, source string) (int, error) {
	now := time.Now().Unix()
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO keywords (word, source, created_at) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	inserted := 0
	for _, w := range words {
		if w == "" {
			continue
		}
		res, err := stmt.Exec(w, source, now)
		if err != nil {
			return 0, err
		}
		n, _ := res.RowsAffected()
		inserted += int(n)
	}
	return inserted, tx.Commit()
}

func (db *DB) CountAvailable() (int, error) {
	cut := time.Now().Add(-10 * 24 * time.Hour).Unix()
	var n int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM keywords WHERE used_at = 0 OR used_at < ?", cut).Scan(&n)
	return n, err
}

func (db *DB) PickRandom(count int) ([]Keyword, error) {
	cut := time.Now().Add(-10 * 24 * time.Hour).Unix()
	rows, err := db.conn.Query(
		"SELECT id, word, source, created_at, used_at FROM keywords WHERE used_at = 0 OR used_at < ? ORDER BY RANDOM() LIMIT ?",
		cut, count,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []Keyword
	for rows.Next() {
		var k Keyword
		if err := rows.Scan(&k.ID, &k.Word, &k.Source, &k.CreatedAt, &k.UsedAt); err != nil {
			return nil, err
		}
		keywords = append(keywords, k)
	}
	return keywords, rows.Err()
}

func (db *DB) MarkUsed(ids []int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	stmt, err := tx.Prepare("UPDATE keywords SET used_at = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.Exec(now, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) DeleteOlderThan(days int) (int, error) {
	cut := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()
	res, err := db.conn.Exec("DELETE FROM keywords WHERE created_at < ?", cut)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (db *DB) DeleteByIDs(ids []int64) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM keywords WHERE id = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	deleted := 0
	for _, id := range ids {
		res, err := stmt.Exec(id)
		if err != nil {
			return 0, err
		}
		n, _ := res.RowsAffected()
		deleted += int(n)
	}
	return deleted, tx.Commit()
}

func (db *DB) GetMeta(key string) (string, error) {
	var val string
	err := db.conn.QueryRow("SELECT value FROM meta WHERE key = ?", key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

func (db *DB) SetMeta(key, value string) error {
	_, err := db.conn.Exec("INSERT OR REPLACE INTO meta (key, value) VALUES (?, ?)", key, value)
	return err
}

type Stats struct {
	Total       int            `json:"total"`
	Available   int            `json:"available"`
	UsedToday   int            `json:"used_today"`
	BySource    map[string]int `json:"by_source"`
	LastCollect string         `json:"last_collect"`
}

func (db *DB) GetStats() (*Stats, error) {
	total, _ := db.Count()
	available, _ := db.CountAvailable()

	todayStart := time.Now().Truncate(24 * time.Hour).Unix()
	var usedToday int
	db.conn.QueryRow("SELECT COUNT(*) FROM keywords WHERE used_at >= ?", todayStart).Scan(&usedToday)

	rows, err := db.conn.Query("SELECT source, COUNT(*) FROM keywords GROUP BY source ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bySource := make(map[string]int)
	for rows.Next() {
		var src string
		var cnt int
		rows.Scan(&src, &cnt)
		if src == "" {
			src = "unknown"
		}
		bySource[src] = cnt
	}

	lastCollect, _ := db.GetMeta("last_collect_time")

	return &Stats{
		Total:       total,
		Available:   available,
		UsedToday:   usedToday,
		BySource:    bySource,
		LastCollect: lastCollect,
	}, nil
}

type KeywordRow struct {
	ID        int64  `json:"id"`
	Word      string `json:"word"`
	Source    string `json:"source"`
	CreatedAt int64  `json:"created_at"`
	UsedAt    int64  `json:"used_at"`
	Status    string `json:"status"`
}

func (db *DB) ListKeywords(page, size int, source, status, search string) ([]KeywordRow, int, error) {
	where := "1=1"
	args := []interface{}{}

	if source != "" {
		where += " AND source LIKE ?"
		args = append(args, source+"%")
	}

	if search != "" {
		where += " AND word LIKE ?"
		args = append(args, "%"+search+"%")
	}

	cut := time.Now().Add(-10 * 24 * time.Hour).Unix()
	switch status {
	case "available":
		where += " AND (used_at = 0 OR used_at < ?)"
		args = append(args, cut)
	case "used":
		where += " AND used_at >= ?"
		args = append(args, cut)
	case "recent":
		sevenDays := time.Now().Add(-7 * 24 * time.Hour).Unix()
		where += " AND created_at >= ?"
		args = append(args, sevenDays)
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM keywords WHERE %s", where)
	if err := db.conn.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	query := fmt.Sprintf("SELECT id, word, source, created_at, used_at FROM keywords WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?", where)
	args = append(args, size, offset)

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var keywords []KeywordRow
	for rows.Next() {
		var k KeywordRow
		if err := rows.Scan(&k.ID, &k.Word, &k.Source, &k.CreatedAt, &k.UsedAt); err != nil {
			return nil, 0, err
		}
		if k.UsedAt == 0 || k.UsedAt < cut {
			k.Status = "available"
		} else {
			k.Status = "used"
		}
		keywords = append(keywords, k)
	}
	return keywords, total, rows.Err()
}

func (db *DB) UpdateLastCollectTime() {
	db.SetMeta("last_collect_time", time.Now().Format(time.RFC3339))
}
package db

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

type Keyword struct {
	ID        int64
	Word      string
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
	return err
}

func (db *DB) Count() (int, error) {
	var n int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM keywords").Scan(&n)
	return n, err
}

func (db *DB) Insert(words []string) (int, error) {
	now := time.Now().Unix()
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO keywords (word, created_at) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	inserted := 0
	for _, w := range words {
		if w == "" {
			continue
		}
		res, err := stmt.Exec(w, now)
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
		"SELECT id, word, created_at, used_at FROM keywords WHERE used_at = 0 OR used_at < ? ORDER BY RANDOM() LIMIT ?",
		cut, count,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []Keyword
	for rows.Next() {
		var k Keyword
		if err := rows.Scan(&k.ID, &k.Word, &k.CreatedAt, &k.UsedAt); err != nil {
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
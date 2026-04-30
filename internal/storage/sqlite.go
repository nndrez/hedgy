package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var _ Repository = (*SQLiteStorage)(nil)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage() (*SQLiteStorage, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("impossible to obtain config dir: %w", err)
	}

	appDir := filepath.Join(configDir, "hedgy")
	os.MkdirAll(appDir, 0755)
	dbPath := filepath.Join(appDir, "hedgy.db")

	dsn := fmt.Sprintf("%s?_journal=WAL&_timeout=5000", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening db: %w", err)
	}

	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE IF NOT EXISTS feeds (
		id 		INTEGER PRIMARY KEY AUTOINCREMENT,
		name 	TEXT NOT NULL,
		url 	TEXT UNIQUE NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS articles (
		id 				INTEGER PRIMARY KEY AUTOINCREMENT,
		feed_id 		INTEGER,
		title 			TEXT NOT NULL,
		link 			TEXT UNIQUE NOT NULL,
		description 	TEXT,
		content 		TEXT,
		published_at 	DATETIME,
		is_read 		BOOLEAN DEFAULT 0,
		FOREIGN KEY(feed_id) REFERENCES feeds(id)
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("error creation tables: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) GetFeeds() ([]Feed, error) {
	rows, err := s.db.Query("SELECT id, name, url FROM feeds")
	if err != nil {
		return nil, fmt.Errorf("error reading feed: %w", err)
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		var f Feed
		if err := rows.Scan(&f.ID, &f.Name, &f.URL); err != nil {
			return nil, fmt.Errorf("error scan feed: %w", err)
		}

		feeds = append(feeds, f)
	}

	return feeds, nil
}

func (s *SQLiteStorage) AddFeed(name, url string) error {
	if _, err := s.db.Exec("INSERT INTO feeds (name, url) VALUES (?, ?)", name, url); err != nil {
		return fmt.Errorf("error insert feed: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) DeleteFeed(id int) error {
	res, err := s.db.Exec("DELETE FROM feeds WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error delete feed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("feed %d not found", id)
	}

	return nil
}

func (s *SQLiteStorage) SaveArticles(feedID int, articles []Article) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error begin transaction: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO articles (feed_id, title, link, description, content, published_at, is_read) 
		VALUES (?, ?, ?, ?, ?, ?, 0) 
		ON CONFLICT(link) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error on prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, a := range articles {
		_, err := stmt.Exec(feedID, a.Title, a.Link, a.Description, a.Content, a.PublishedAt)
		if err != nil {
			return fmt.Errorf("error saving article: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) GetUnreadArticles(feedID, limit int) ([]Article, error) {
	rows, err := s.db.Query(`
		SELECT id, feed_id, title, link, description, content, published_at, is_read 
		FROM articles 
		WHERE feed_id = ? AND is_read = 0 
		ORDER BY published_at DESC 
		LIMIT ?`, feedID, limit)

	if err != nil {
		return nil, fmt.Errorf("error fetching unread articles: %w", err)
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.Link, &a.Description, &a.Content, &a.PublishedAt, &a.IsRead); err != nil {
			return nil, fmt.Errorf("error during article scan: %w", err)
		}

		articles = append(articles, a)
	}

	return articles, nil
}

func (s *SQLiteStorage) MarkAsRead(articleID int) error {
	_, err := s.db.Exec("UPDATE articles SET is_read = 1 WHERE id = ?", articleID)
	return err
}

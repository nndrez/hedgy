package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Feed struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Repository interface {
	GetFeeds() ([]Feed, error)
	AddFeed(name, url string) error
	DeleteFeed(id int) error
}

type JSONStorage struct {
	filePath string
	mu       sync.RWMutex
}

func NewJSONStorage() (*JSONStorage, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("impossible obtain config dir: %w", err)
	}

	appDir := filepath.Join(configDir, "hedgy")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("impossible to create hedgy dir: %w", err)
	}

	filePath := filepath.Join(appDir, "feeds.json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.WriteFile(filePath, []byte("[]"), 0644); err != nil {
			return nil, fmt.Errorf("feeds.json initialization failure: %w", err)
		}
	}

	return &JSONStorage{filePath: filePath}, nil
}

func (s *JSONStorage) GetFeeds() ([]Feed, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var feeds []Feed
	if err := json.Unmarshal(data, &feeds); err != nil {
		return nil, fmt.Errorf("error deserializing data: %w", err)
	}

	return feeds, nil
}

func (s *JSONStorage) AddFeed(name, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("reading file error before adding: %w", err)
	}

	var feeds []Feed
	if err := json.Unmarshal(data, &feeds); err != nil {
		return fmt.Errorf("error deserializing data: %w", err)
	}

	// simple ID max calculus
	maxID := 0
	for _, f := range feeds {
		if f.ID > maxID {
			maxID = f.ID
		}
	}
	newID := maxID + 1

	feeds = append(feeds, Feed{ID: newID, Name: name, URL: url})

	return s.save(feeds)
}

func (s *JSONStorage) DeleteFeed(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("reading file error: %w", err)
	}

	var feeds []Feed
	if err := json.Unmarshal(data, &feeds); err != nil {
		return fmt.Errorf("error deserializing data: %w", err)
	}

	var newFeeds []Feed
	found := false
	for _, f := range feeds {
		if f.ID == id {
			found = true
			continue
		}
		newFeeds = append(newFeeds, f)
	}

	if !found {
		return fmt.Errorf("not found feed with %d ID", id)
	}

	return s.save(newFeeds)
}

func (s *JSONStorage) save(feeds []Feed) error {
	data, err := json.MarshalIndent(feeds, "", " ")
	if err != nil {
		return fmt.Errorf("error serializing data: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("error saving: %w", err)
	}

	return nil
}

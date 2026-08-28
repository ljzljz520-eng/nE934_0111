package store

import (
	"errors"
	"fmt"
	"go.etcd.io/bbolt"
	"path/filepath"
	"sync"
)

var ErrNotFound = errors.New("record not found")

type Store struct {
	db   *bbolt.DB
	mu   sync.RWMutex
	path string
}

var bucketNames = [][]byte{[]byte("records"), []byte("audits"), []byte("workflows"), []byte("attachments")}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	db, err := bbolt.Open(filepath.Clean(path), 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	s := &Store{db: db, path: path}
	if err := s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range bucketNames {
			if _, e := tx.CreateBucketIfNotExists(name); e != nil {
				return e
			}
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Path() string { s.mu.RLock(); defer s.mu.RUnlock(); return s.path }

func (s *Store) Update(fn func(*bbolt.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store closed")
	}
	return s.db.Update(fn)
}

func (s *Store) View(fn func(*bbolt.Tx) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store closed")
	}
	return s.db.View(fn)
}

func cloneBytes(v []byte) []byte {
	if v == nil {
		return nil
	}
	out := make([]byte, len(v))
	copy(out, v)
	return out
}

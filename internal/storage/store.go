package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

const (
	appDirName = "tld"
	dbFileName = "tld.db"
)

var ErrEmptyPassphrase = errors.New("database passphrase is empty")

type Store struct {
	db   *sql.DB
	path string
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}

	return filepath.Join(configDir, appDirName, dbFileName), nil
}

func Open(ctx context.Context, path string, passphrase string) (*Store, error) {
	if passphrase == "" {
		return nil, ErrEmptyPassphrase
	}

	if path == "" {
		defaultPath, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		path = defaultPath
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", buildDSN(path, passphrase))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db, path: path}
	if err := store.verify(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := Migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	_ = os.Chmod(path, 0o600)

	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.Close()
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}

	return s.path
}

func (s *Store) Cache() CacheRepository {
	return CacheRepository{db: s.db}
}

func (s *Store) Stacks() StackRepository {
	return StackRepository{db: s.db}
}

func (s *Store) Namespaces() NamespaceRepository {
	return NamespaceRepository{db: s.db}
}

func (s *Store) Dependencies() DependencyRepository {
	return DependencyRepository{db: s.db}
}

func (s *Store) Sources() SourceRepository {
	return SourceRepository{db: s.db}
}

func (s *Store) Policies() PolicyRepository {
	return PolicyRepository{db: s.db}
}

func (s *Store) verify(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master").Scan(&count); err != nil {
		return fmt.Errorf("unlock database: %w", err)
	}

	return nil
}

func buildDSN(path string, passphrase string) string {
	values := url.Values{}
	values.Set("_busy_timeout", "5000")
	values.Set("_foreign_keys", "on")
	values.Set("_pragma_cipher_page_size", "4096")
	values.Set("_pragma_kdf_iter", "256000")
	values.Set("_pragma_key", passphrase)

	dsn := url.URL{Scheme: "file", Path: path, RawQuery: values.Encode()}
	return dsn.String()
}

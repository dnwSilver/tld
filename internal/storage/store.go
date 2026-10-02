package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

const (
	appDirName         = "tld"
	dbFileName         = "tld.db"
	cachePayloadBudget = 256 << 20
	databaseKDFIter    = 256000
	databasePageSize   = 4096
)

var ErrEmptyPassphrase = errors.New("database passphrase is empty")

type Store struct {
	db              *sql.DB
	path            string
	migrationBackup string
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

	defaultLocation := path == ""
	if defaultLocation {
		defaultPath, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		path = defaultPath
	}

	_, statErr := os.Stat(path)
	existed := statErr == nil
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect database: %w", statErr)
	}
	directory := filepath.Dir(path)
	_, directoryStatErr := os.Stat(directory)
	directoryCreated := errors.Is(directoryStatErr, os.ErrNotExist)
	if directoryStatErr != nil && !directoryCreated {
		return nil, fmt.Errorf("inspect database directory: %w", directoryStatErr)
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	// Never change an arbitrary existing parent supplied by the caller (for
	// example /tmp or the current working directory). The app-owned default
	// directory and directories created here are safe to tighten.
	if defaultLocation || directoryCreated {
		if err := os.Chmod(directory, 0o700); err != nil {
			return nil, fmt.Errorf("secure database directory: %w", err)
		}
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
	if err := store.verifyCipherSettings(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if existed {
		version, err := databaseSchemaVersion(ctx, db)
		if err != nil {
			_ = db.Close()
			return nil, err
		}
		if version < currentSchemaVersion {
			backupPath := fmt.Sprintf("%s.bak-v%d-%s", path, version, time.Now().UTC().Format("20060102T150405Z"))
			if err := copyDatabaseFile(path, backupPath); err != nil {
				_ = db.Close()
				return nil, err
			}
			store.migrationBackup = backupPath
		}
	}

	if err := Migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.Cache().DeleteAmbiguousProjectKeys(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.Cache().DeleteExpired(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.Cache().EnforceByteBudget(ctx, cachePayloadBudget); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := os.Chmod(path, 0o600); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("secure database file: %w", err)
	}

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

func (s *Store) MigrationBackupPath() string {
	if s == nil {
		return ""
	}
	return s.migrationBackup
}

func (s *Store) Cache() CacheRepository {
	return CacheRepository{db: s.db, maxBytes: cachePayloadBudget}
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

func (s *Store) Projects() ProjectRepository {
	return ProjectRepository{db: s.db}
}

func (s *Store) ProjectDependencies() ProjectDependencyRepository {
	return ProjectDependencyRepository{db: s.db}
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
	values.Set("_pragma_cipher_page_size", fmt.Sprint(databasePageSize))
	values.Set("_pragma_kdf_iter", fmt.Sprint(databaseKDFIter))
	// The driver wraps _pragma_key in a double-quoted PRAGMA literal without
	// escaping it. Double quotes must therefore be doubled before URL encoding.
	values.Set("_pragma_key", strings.ReplaceAll(passphrase, `"`, `""`))

	dsn := url.URL{Scheme: "file", Path: path, RawQuery: values.Encode()}
	return dsn.String()
}

func (s *Store) verifyCipherSettings(ctx context.Context) error {
	var kdfIter int
	if err := s.db.QueryRowContext(ctx, "PRAGMA kdf_iter").Scan(&kdfIter); err != nil {
		return fmt.Errorf("read SQLCipher kdf_iter: %w", err)
	}
	if kdfIter != databaseKDFIter {
		return fmt.Errorf("SQLCipher kdf_iter = %d, want %d", kdfIter, databaseKDFIter)
	}
	var pageSize int
	if err := s.db.QueryRowContext(ctx, "PRAGMA cipher_page_size").Scan(&pageSize); err != nil {
		return fmt.Errorf("read SQLCipher page size: %w", err)
	}
	if pageSize != databasePageSize {
		return fmt.Errorf("SQLCipher page size = %d, want %d", pageSize, databasePageSize)
	}
	return nil
}

func databaseSchemaVersion(ctx context.Context, db *sql.DB) (int, error) {
	var exists int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'").Scan(&exists); err != nil {
		return 0, fmt.Errorf("inspect schema migrations: %w", err)
	}
	if exists == 0 {
		return 0, nil
	}
	var version sql.NullInt64
	if err := db.QueryRowContext(ctx, "SELECT max(version) FROM schema_migrations").Scan(&version); err != nil {
		return 0, fmt.Errorf("read database schema version: %w", err)
	}
	if !version.Valid {
		return 0, nil
	}
	return int(version.Int64), nil
}

func copyDatabaseFile(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open database backup source: %w", err)
	}
	defer source.Close()
	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create database migration backup: %w", err)
	}
	remove := true
	defer func() {
		_ = target.Close()
		if remove {
			_ = os.Remove(targetPath)
		}
	}()
	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("copy database migration backup: %w", err)
	}
	if err := target.Sync(); err != nil {
		return fmt.Errorf("sync database migration backup: %w", err)
	}
	if err := target.Close(); err != nil {
		return fmt.Errorf("close database migration backup: %w", err)
	}
	remove = false
	return nil
}

package data

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	grpcv1migration "web-api/gen/sqlite/migrations"
	"web-api/internal/core"
)

// StorageManager はデータストレージサービスを管理します。
type StorageManager struct {
	PersisterMap map[string]*Storager
	db           *sql.DB
	migrations   []string
}

// NewStorageManager は StorageManager の新しいインスタンスを作成します。
func NewStorageManager() *StorageManager {
	srv := &StorageManager{}
	srv.PersisterMap = make(map[string]*Storager)
	srv.migrations = grpcv1migration.DefaultMigrations()
	return srv
}

// Storager はファイルシステムとデータベースを橋渡しするインターフェースを定義します。
type Storager interface {
	Name() string
	Start(*StorageManager) error
	SyncToDB(*sql.DB) error
	Cleanup()
}

// AddStorager はサービスを追加する
func (sm *StorageManager) AddStorager(storager Storager) {
	sm.PersisterMap[storager.Name()] = &storager
}

// Start はすべてのサービスを起動する
func (sm *StorageManager) Start() error {
	// データベース接続
	if err := sm.openDB(); err != nil {
		return err
	}
	if err := sm.applyMigrations(); err != nil {
		return err
	}
	for _, p := range sm.PersisterMap {
		if err := (*p).Start(sm); err != nil {
			return err
		}
		if err := (*p).SyncToDB(sm.db); err != nil {
			return err
		}
	}
	return nil
}

// CleanupAll はサービスをクリーンアップする
func (sm *StorageManager) CleanupAll() {
	for _, srv := range sm.PersisterMap {
		(*srv).Cleanup()
	}
	if sm.db != nil {
		_ = sm.db.Close()
	}
}

// openDB はデータベース接続を開く
func (sm *StorageManager) openDB() error {
	if sm.db != nil {
		return nil
	}
	dbPath := core.Config.PersistDBPath
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	sm.db = db
	return nil
}

func (sm *StorageManager) applyMigrations() error {
	if len(sm.migrations) == 0 {
		return nil
	}
	for _, query := range sm.migrations {
		if _, err := sm.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (sm *StorageManager) DB() *sql.DB {
	return sm.db
}

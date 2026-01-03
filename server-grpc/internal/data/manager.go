package data

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	grpcv1migration "server-grpc/gen/sqlite/migrations"
	"server-grpc/internal/core"
)

// StorageManager は各サービスのハンドラーをまとめた構造体です。
type StorageManager struct {
	PersisterMap map[string]*Storage
	db           *sql.DB
	migrations   []string
}

// NewPersistManager は与えられたオプションでサービス群を初期化します。
func NewPersistManager() *StorageManager {
	srv := &StorageManager{}
	srv.PersisterMap = make(map[string]*Storage)
	srv.migrations = grpcv1migration.DefaultMigrations()
	return srv
}

// Storage はファイルシステムとデータベースを橋渡しするインターフェースを定義します。
type Storage interface {
	Name() string
	Start(*StorageManager) error
	SyncToDB(*sql.DB) error
	Cleanup()
}

// AddPersister はサービスを追加する
func (sm *StorageManager) AddPersister(persister Storage) {
	sm.PersisterMap[persister.Name()] = &persister
}

// Start はすべてのサービスを起動する
func (sm *StorageManager) Start() error {
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

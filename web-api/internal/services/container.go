package services

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	grpcv1migration "web-api/gen/sqlite/migrations"
	"web-api/internal/core"
)

// ServiceContainer はデータストレージサービスを管理します。
type ServiceContainer struct {
	db         *sql.DB
	migrations []string
	Services   map[string]*Service
}

// NewServiceContainer は ServiceContainer の新しいインスタンスを作成します。
func NewServiceContainer() *ServiceContainer {
	srv := &ServiceContainer{}
	srv.Services = make(map[string]*Service)
	srv.migrations = grpcv1migration.DefaultMigrations()
	return srv
}

// Service はファイルシステムとデータベースを橋渡しするインターフェースを定義します。
type Service interface {
	Name() string
	Start(*ServiceContainer) error
	SyncToDB(*sql.DB) error
	Cleanup()
}

// AddService はサービスを追加する
func (sm *ServiceContainer) AddService(service Service) {
	sm.Services[service.Name()] = &service
}

// Start はすべてのサービスを起動する
func (sm *ServiceContainer) Start() error {
	// データベース接続
	if err := sm.openDB(); err != nil {
		return err
	}
	if err := sm.applyMigrations(); err != nil {
		return err
	}
	for _, p := range sm.Services {
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
func (sm *ServiceContainer) CleanupAll() {
	for _, srv := range sm.Services {
		(*srv).Cleanup()
	}
	if sm.db != nil {
		_ = sm.db.Close()
	}
}

// openDB はデータベース接続を開く
func (sm *ServiceContainer) openDB() error {
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

func (sm *ServiceContainer) applyMigrations() error {
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

func (sm *ServiceContainer) DB() *sql.DB {
	return sm.db
}

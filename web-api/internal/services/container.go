package ctrl

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	grpcv1migration "web-api/gen/sqlite/migrations"
	"web-api/internal/core"
)

// ContainerService はデータストレージサービスを管理します。
type ContainerService struct {
	PersisterMap map[string]*Service
	db           *sql.DB
	migrations   []string
}

// NewContainerService は ContainerService の新しいインスタンスを作成します。
func NewContainerService() *ContainerService {
	srv := &ContainerService{}
	srv.PersisterMap = make(map[string]*Service)
	srv.migrations = grpcv1migration.DefaultMigrations()
	return srv
}

// Service はファイルシステムとデータベースを橋渡しするインターフェースを定義します。
type Service interface {
	Name() string
	Start(*ContainerService) error
	SyncToDB(*sql.DB) error
	Cleanup()
}

// AddService はサービスを追加する
func (cs *ContainerService) AddService(service Service) {
	cs.PersisterMap[service.Name()] = &service
}

// Start はすべてのサービスを起動する
func (cs *ContainerService) Start() error {
	// データベース接続
	if err := cs.openDB(); err != nil {
		return err
	}
	if err := cs.applyMigrations(); err != nil {
		return err
	}
	for _, p := range cs.PersisterMap {
		if err := (*p).Start(cs); err != nil {
			return err
		}
		if err := (*p).SyncToDB(cs.db); err != nil {
			return err
		}
	}
	return nil
}

// CleanupAll はサービスをクリーンアップする
func (cs *ContainerService) CleanupAll() {
	for _, srv := range cs.PersisterMap {
		(*srv).Cleanup()
	}
	if cs.db != nil {
		_ = cs.db.Close()
	}
}

// openDB はデータベース接続を開く
func (cs *ContainerService) openDB() error {
	if cs.db != nil {
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
	cs.db = db
	return nil
}

func (cs *ContainerService) applyMigrations() error {
	if len(cs.migrations) == 0 {
		return nil
	}
	for _, query := range cs.migrations {
		if _, err := cs.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (cs *ContainerService) DB() *sql.DB {
	return cs.db
}

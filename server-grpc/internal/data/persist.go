package data

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	grpcv1migration "server-grpc/gen/sqlite/migrations"
	"server-grpc/internal/core"
)

// PersistHub は各サービスのハンドラーをまとめた構造体です。
type PersistHub struct {
	ServiceMap map[string]*PersistBridge
	db         *sql.DB
	migrations []string
}

// NewPersistHub は与えられたオプションでサービス群を初期化します。
func NewPersistHub() *PersistHub {
	srv := &PersistHub{}
	srv.ServiceMap = make(map[string]*PersistBridge)
	srv.migrations = grpcv1migration.DefaultMigrations()
	return srv
}

// PersistBridge はファイルシステムとデータベースを橋渡しするインターフェースを定義します。
type PersistBridge interface {
	Name() string
	Start(*PersistHub, *map[string]string) error
	SyncToDB(*sql.DB) error
	Cleanup()
}

// AddService はサービスを追加する
func (ss *PersistHub) AddService(service PersistBridge) {
	ss.ServiceMap[service.Name()] = &service
}

// StartAll はすべてのサービスを起動する
func (ss *PersistHub) StartAll() error {
	if err := ss.openDB(); err != nil {
		return err
	}
	if err := ss.applyMigrations(); err != nil {
		return err
	}
	for _, s := range ss.ServiceMap {
		if err := (*s).Start(ss, &core.ServerConfiguration); err != nil {
			return err
		}
		if err := (*s).SyncToDB(ss.db); err != nil {
			return err
		}
	}
	return nil
}

// CleanupAll はサービスをクリーンアップする
func (ss *PersistHub) CleanupAll() {
	for _, srv := range ss.ServiceMap {
		(*srv).Cleanup()
	}
	if ss.db != nil {
		_ = ss.db.Close()
	}
}

func (ss *PersistHub) openDB() error {
	if ss.db != nil {
		return nil
	}
	dbPath := core.ServerConfiguration["PersistDBPath"]
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	ss.db = db
	return nil
}

func (ss *PersistHub) applyMigrations() error {
	if len(ss.migrations) == 0 {
		return nil
	}
	for _, query := range ss.migrations {
		if _, err := ss.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (ss *PersistHub) DB() *sql.DB {
	return ss.db
}

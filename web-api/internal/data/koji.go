package data

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"web-api/internal/core"
	"sync"

	grpcv1 "web-api/gen/grpc/v1"
	grpcv1connect "web-api/gen/grpc/v1/grpcv1connect"
	"web-api/internal/models"

	"connectrpc.com/connect"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KojiStorage bridges existing KojiStorage logic to Connect handlers.
type KojiStorage struct {
	// Embed the unimplemented handler for forward compatibility
	grpcv1connect.UnimplementedKojiServiceHandler

	// manager は任意のgrpcサービスハンドラーへの参照
	manager *StorageManager

	// folderPath はこのサービスが管理する工事データのルートフォルダー
	folderPath string

	// watcher は target のファイルシステム監視オブジェクト
	watcher *fsnotify.Watcher
}

// Name はサービス名を返します
func (s *KojiStorage) Name() string {
	return "KojiService"
}

func (s *KojiStorage) Start(sm *StorageManager) error {
	// パスの取得と正規化
	target, err := core.NormalizeAbsPath(core.Config.KojiTargetFolder)
	if err != nil {
		return err
	}

	// 情報の初期化
	s.manager = sm
	s.folderPath = target

	// kojiesByIdの情報を取得
	if err = s.UpdateKojies(); err != nil {
		return err
	}

	// targetの監視を開始
	if err = s.watchTarget(); err != nil {
		return err
	}

	return nil
}

func (s *KojiStorage) Cleanup() {
	// 現在はクリーンアップ処理は不要
}

// SyncToDB は工事データを SQLite に同期する。
func (s *KojiStorage) SyncToDB(db *sql.DB) error {
	return s.persistKojies(db)
}

// watchTarget starts watching the provided target for changes.
// Add callbacks or channels as needed to propagate events to your services.
func (s *KojiStorage) watchTarget() error {
	absPath, err := filepath.Abs(s.folderPath)
	if err != nil {
		return err
	}

	s.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	// 監視終了時に閉じる
	go func() {
		<-s.watcher.Errors
		s.watcher.Close()
	}()

	// イベントループ
	go func() {
		for {
			select {
			case event, ok := <-s.watcher.Events:
				if !ok {
					return
				}
				log.Printf("[target] event=%s path=%s", event.Op, event.Name)
				if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Rename|fsnotify.Write) != 0 {
					// 必要に応じてサービスへ通知する
					// 例: reload metadata, update cache, etc.
				}

			case err := <-s.watcher.Errors:
				log.Printf("[target] watcher error: %v", err)
			}
		}
	}()

	// フォルダを監視対象に追加
	if err := s.watcher.Add(absPath); err != nil {
		return err
	}

	log.Printf("watching target: %s", absPath)
	return nil
}

func (s *KojiStorage) UpdateKojies() error {
	// ファイルシステムから工事フォルダー一覧を取得
	entries, err := os.ReadDir(s.folderPath)
	if err != nil {
		return err
	}

	// 工事フォルダー一覧の要素数を取得
	kojiesSize := len(entries)

	// 並列処理用のワーカー数を決定
	numWorkers := core.DecideNumWorkers(kojiesSize)

	// バッファ付きチャンネルで効率化
	jobs := make(chan int, kojiesSize)
	results := make(chan *models.Koji, kojiesSize)

	// ワーカープールの起動
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for idx := range jobs {
				folder := path.Join(s.folderPath, entries[idx].Name())
				koji := models.NewKoji()
				if err := koji.ParseFrom(folder); err == nil {
					results <- koji
				} else {
					results <- nil // エラーの場合はnilを返す
				}
			}
		}()
	}

	// ジョブの投入
	go func() {
		for i := range entries {
			jobs <- i
		}
		close(jobs)
	}()

	// 結果収集用のゴルーチン
	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result != nil {
			s.kojies[result.GetId()] = result
		}
	}

	if s.manager != nil {
		return s.persistKojies(s.manager.DB())
	}

	return nil
}

// GetKojies は管理されている工事データ一覧を返す
func (s *KojiStorage) GetKojies(
	ctx context.Context,
	req *grpcv1.GetKojiesRequest) (
	res *grpcv1.GetKojiesResponse,
	err error) {
	_ = req // 現状フィルター未対応

	grpcKojies := make(map[string]*grpcv1.Koji, len(s.kojies))
	for _, v := range s.kojies {
		grpcKojies[v.GetId()] = v.Koji
	}

	res.SetKojies(grpcKojies)

	return
}

// GetKojiById は指定されたIDの工事データを返す
func (s *KojiStorage) GetKojiById(
	ctx context.Context,
	req *grpcv1.GetKojiRequest) (
	res *grpcv1.GetKojiResponse,
	err error) {

	// リクエスト情報の取得
	id := req.GetId()

	// 工事情報を取得
	koji, exist := s.kojies[id]
	if !exist {
		err = connect.NewError(connect.CodeNotFound, errors.New("koji not found"))
		return
	}

	// Responseの更新
	res.SetKoji(koji.Koji)

	return
}

func (s *KojiStorage) UpdateKoji(
	_ context.Context, req *grpcv1.UpdateKojiRequest) (
	*grpcv1.UpdateKojiResponse, error) {

	// 既存の工事情報を取得
	grpcNewKoji := req.GetNewKoji()
	prevKoji, exist := s.kojies[grpcNewKoji.GetId()]
	if !exist {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("koji not found"))
	}

	newKoji := &models.Koji{Koji: grpcNewKoji}

	// 工事情報を更新
	newKoji, err := prevKoji.ImportFrom(newKoji)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// 工事情報のインデックスを更新
	if _, exist := s.kojies[prevKoji.GetId()]; exist {
		delete(s.kojies, prevKoji.GetId())
		// 新しいIDで再登録
		s.kojies[newKoji.GetId()] = newKoji
	}

	// Responseの作成
	res := grpcv1.UpdateKojiResponse_builder{}.Build()

	grpcv1KojiMapById := make(map[string]*grpcv1.Koji, len(s.kojies))
	for _, v := range s.kojies {
		grpcv1KojiMapById[v.GetId()] = v.Koji
	}
	res.SetPrevKoji(prevKoji.Koji)

	return res, nil
}

func (s *KojiStorage) persistKojies(db *sql.DB) error {
	if db == nil {
		return errors.New("koji persist db is nil")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM kojies`); err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO kojies (id, status, pathist_folder, start_at, company_name, location_name, end_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, koji := range s.kojies {
		if _, err := stmt.Exec(
			koji.GetId(),
			koji.GetStatus(),
			koji.GetPathistFolder(),
			timestampValue(koji.GetStart()),
			koji.GetCompanyName(),
			koji.GetLocationName(),
			timestampValue(koji.GetPersistEnd()),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func timestampValue(ts *timestamppb.Timestamp) any {
	if ts == nil || !ts.IsValid() {
		return nil
	}
	return ts.AsTime()
}

// RenameStandardFile は標準ファイルの名前を変更し、工事データも更新する
// TODO: StandardFile型が定義されていないため、一時的にコメントアウト
// func (ks *KojiService) RenameStandardFile(koji models.Koji, actuals []string) []string {
// 	// マップの作成
// 	actualToStandardMap := make(map[string]*models.StandardFile)
// 	for i := range koji.StandardFiles {
// 		sf := &koji.StandardFiles[i]
// 		actualToStandardMap[sf.ActualName] = sf
// 	}
//
// 	// 変更後の標準ファイル名を格納する配列
// 	renamedFiles := make([]string, len(actuals))
//
// 	// 変更前の標準ファイル名をループ
// 	count := 0
// 	for _, actual := range actuals {
// 		if sf, exists := actualToStandardMap[actual]; exists {
// 			actualFullpath, err := ks.BaseFolderService.GetFullpath(sf.GetPath())
// 			if err != nil {
// 				continue
// 			}
//
// 			standardFullpath, err := ks.BaseFolderService.GetFullpath(sf.Name)
// 			if err != nil {
// 				continue
// 			}
// 			count++
// 		}
//
// 		// ファイル名変更後、工事の必須ファイル情報を更新
// 		if count > 0 {
// 			// 必須ファイル情報を再設定
// 			err := ks.UpdateRequiredFiles(&koji)
// 			if err == nil {
// 				// 属性ファイルに反映
// 				ks.DatabaseService.Save(&koji)
// 			}
// 		}
// 	}
//
// 	// ファイル名変更後、工事の必須ファイル情報を更新
// 	if count > 0 {
// 		// 必須ファイル情報を再設定
// 		err := ks.UpdateRequiredFiles(&koji)
// 		if err == nil {
// 			// 属性ファイルに反映
// 			ks.DatabaseService.Save(&koji)
// 		}
// 	}
//
// 	return renamedFiles[:count]
// }

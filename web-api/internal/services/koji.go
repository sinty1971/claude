package services

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"web-api/internal/core"

	grpcv1 "web-api/gen/grpc/v1"
	grpcv1connect "web-api/gen/grpc/v1/grpcv1connect"
	"web-api/internal/models"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KojiService bridges existing KojiService logic to Connect handlers.
type KojiService struct {
	// CS は任意のgrpcサービスハンドラーへの参照
	CS *ContainerService

	// DirPath はこのサービスが管理する工事データのルートフォルダー
	DirPath string

	// cache は工事データのIDをキーとしたキャッシュマップ
	cache map[string]*models.Koji

	// watcher はファイルシステム監視オブジェクト
	watcher *core.Watcher

	// Embed the unimplemented handler for forward compatibility
	grpcv1connect.UnimplementedKojiServiceHandler
}

// Name はサービス名を返します
func (srv *KojiService) Name() string {
	return "KojiService"
}

func (srv *KojiService) Start(cs *ContainerService) error {
	// パスの取得と正規化
	dirPath, err := core.NormalizeAbsPath(core.Config.KojiServiceDirPath)
	if err != nil {
		return err
	}

	// 情報の初期化
	srv.CS = cs
	srv.DirPath = dirPath
	srv.cache = make(map[string]*models.Koji)

	// kojiesByIdの情報を取得
	if err = srv.SyncAllToCache(); err != nil {
		return err
	}

	// ファイルシステム監視オブジェクトの作成
	watcher, err := core.NewWatcher()
	if err != nil {
		return err
	}
	srv.watcher = watcher

	// 監視対象ディレクトリの設定
	err = srv.watcher.Start(srv.DirPath, core.Config.CompanyWatcherMaxDepth)
	if err != nil {
		return err
	}

	// ゴルーチンで監視イベントを処理
	go srv.consumeWatcherEvents()

	return nil
}

func (srv *KojiService) Cleanup() {
	// 現在はクリーンアップ処理は不要
}

// consumeWatcherEvents はファイルシステム監視イベントを処理します
func (srv *KojiService) consumeWatcherEvents() {
	for {
		select {
		case event, ok := <-srv.watcher.Events():
			if !ok {
				return
			}
			log.Printf("KojiService: File system event: %s", event)

			// データベースへの同期
			if err := srv.SyncAllToCache(); err != nil {
				log.Printf("KojiService: Failed to update koji cache map: %v", err)
			}

		case err, ok := <-srv.watcher.Errors():
			if !ok {
				return
			}
			log.Printf("KojiService: File system watcher error: %v", err)
		}
	}
}

func (srv *KojiService) SyncAllToCache() error {
	// ファイルシステムから工事フォルダー一覧を取得
	entries, err := os.ReadDir(srv.DirPath)
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
				dirPath := filepath.Join(srv.DirPath, entries[idx].Name())
				koji, err := models.NewKoji(dirPath)
				if err == nil {
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
			srv.cache[result.GetId()] = result
		}
	}

	return nil
}

// Update は指定 targetId のキャッシュ情報を新しい会社情報で更新します
//
//	targetId: 更新対象会社Id
//	source: 新しい会社情報
func (srv *KojiService) Update(targetId string, source *models.Koji) error {
	// 引数チェック
	if source == nil {
		return errors.New("更新情報 source の値が nil です")
	}

	// targetId から工事データを取得
	target, exist := srv.cache[targetId]
	if !exist {
		return errors.New("更新対象の工事情報が存在しません")
	}

	// 工事情報の更新
	err := target.Update(source)
	if err != nil {
		return err
	}

	// キャッシュ情報の更新
	srv.cache[target.GetId()] = target

	return nil
}

// GetKojies は管理されている工事データ一覧を返す
func (srv *KojiService) GetKojies(
	ctx context.Context,
	req *grpcv1.GetKojiesRequest) (
	res *grpcv1.GetKojiesResponse,
	err error) {
	_ = req // 現状フィルター未対応

	grpcKojies := make(map[string]*grpcv1.Koji, len(srv.cache))
	for _, v := range srv.cache {
		grpcKojies[v.GetId()] = v.Koji
	}

	res.SetKojies(grpcKojies)

	return
}

// GetKojiById は指定されたIDの工事データを返す
func (srv *KojiService) GetKoji(
	// args
	ctx context.Context,
	req *grpcv1.GetKojiRequest) (

	// returns
	res *grpcv1.GetKojiResponse,
	err error) {

	// リクエスト情報の取得
	id := req.GetTargetId()

	// 工事情報を取得
	koji, exist := srv.cache[id]
	if !exist {
		err = connect.NewError(connect.CodeNotFound, errors.New("koji not found"))
		return
	}

	// Responseの更新
	res.SetKoji(koji.Koji)

	return
}

func (srv *KojiService) UpdateKoji(
	// args
	ctx context.Context,
	req *grpcv1.UpdateKojiRequest) (

	// returns
	res *grpcv1.UpdateKojiResponse,
	err error) {

	// 既存の工事情報を取得
	targetId := req.GetTargetId()
	target, exist := srv.cache[targetId]
	if !exist {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("koji not found"))
	}

	// 変更前の工事データのメッセージを保存
	prevMessageKoji := proto.Clone(target.Koji).(*grpcv1.Koji)

	newKoji, err := models.NewKojiFromMessage(req.GetSourceKoji())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// 工事情報を更新
	err = srv.Update(targetId, newKoji)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// Responseの作成
	res = grpcv1.UpdateKojiResponse_builder{}.Build()
	res.SetPrevKoji(prevMessageKoji)

	return res, nil
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

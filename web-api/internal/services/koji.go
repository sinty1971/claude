package services

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
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
	// name はサービス名
	name string

	// cs は任意のgrpcサービスハンドラーへの参照
	cs *ContainerService

	// baseDirPath はこのサービスが管理する工事一覧のディレクトリパス
	baseDirPath string

	// repo は工事データのリポジトリ（自動保存有効）
	repo *core.Repository[*grpcv1.Koji, *models.Koji]

	// watcher はファイルシステム監視オブジェクト
	watcher *core.Watcher

	// Embed the unimplemented handler for forward compatibility
	grpcv1connect.UnimplementedKojiServiceHandler
}

func NewKojiService(cs *ContainerService) *KojiService {
	// パスをの取得と正規化
	baseDirPath, err := core.ResolveAbsPath(core.Config.KojiBaseDirPath)
	if err != nil {
		panic(err)
	}

	return &KojiService{
		name:        "KojiService",
		baseDirPath: baseDirPath,
		repo:        core.NewRepository[*grpcv1.Koji, *models.Koji](true), // 自動保存有効
		cs:          cs,
	}
}

// Name はサービス名を返します
func (srv *KojiService) Name() string {
	return "KojiService"
}

// GenerateHandler はサービスのハンドラを生成します
func (srv *KojiService) GenerateHandler() (
	servicePath string, handler http.Handler, serviceName string) {

	// gRPC パスとハンドラの生成
	servicePath, handler = grpcv1connect.NewKojiServiceHandler(srv)

	// サービス名の取得
	serviceName = grpcv1connect.KojiServiceName

	return
}

func (srv *KojiService) Start() error {
	// キャッシュマップを初期化
	err := srv.SyncAllToCache()
	if err != nil {
		return err
	}

	// watcher の作成と初期化
	watcher, err := core.NewWatcher(srv.baseDirPath, core.Config.KojiWatcherMaxDepth)
	if err != nil {
		panic(err)

	}
	srv.watcher = watcher

	// 監視対象ディレクトリの設定
	err = srv.watcher.Start()
	if err != nil {
		return err
	}

	// ゴルーチンで監視イベントを処理
	go srv.watchFileSystemEvents()

	return nil
}

func (srv *KojiService) Cleanup() {
	// 現在はクリーンアップ処理は不要
}

// watchFileSystemEvents はファイルシステム監視イベントを処理します
func (srv *KojiService) watchFileSystemEvents() {
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
	// Repositoryをクリア
	srv.repo.Clear()

	// ファイルシステムから工事フォルダー一覧を取得
	entries, err := os.ReadDir(srv.baseDirPath)
	if err != nil {
		return err
	}

	// 工事フォルダー一覧の要素数を取得
	entriesCount := len(entries)

	// 並列処理用のワーカー数を決定
	workderCount := core.Config.CalculateWorkerCount(entriesCount)

	// バッファ付きチャンネルで効率化
	chanCount := make(chan int, entriesCount)
	chanKojies := make(chan *core.PersistModel[*grpcv1.Koji, *models.Koji], entriesCount)

	// ワーカープールの起動
	var wg sync.WaitGroup
	wg.Add(workderCount)
	for range workderCount {
		go func() {
			defer wg.Done()
			for i := range chanCount {
				dirPath := core.PathJoin(srv.baseDirPath, entries[i].Name())
				koji, err := models.NewPersistModelKoji(dirPath)
				if err != nil {
					chanKojies <- nil // エラーの場合はnilを返す
					continue
				}

				err = koji.Load() // マニフェストの読み込み
				if err != nil {
					chanKojies <- nil // エラーの場合はnilを返す
					continue
				}

				chanKojies <- koji
			}
		}()
	}

	// ジョブの投入
	go func() {
		for i := range entries {
			chanCount <- i
		}
		close(chanCount)
	}()

	// 結果収集用のゴルーチン
	go func() {
		wg.Wait()
		close(chanKojies)
	}()

	// キャッシュマップの更新
	srv.repo.Clear()
	srv.repo.SetAutoSave(false)
	for koji := range chanKojies {
		if koji != nil {
			if err := srv.repo.Set(*koji); err != nil {
				log.Printf("リポジトリへの追加に失敗しました: %v", err)
			}
		}
	}
	srv.repo.SetAutoSave(true)

	return nil
}

// GetKojies は管理されている工事データ一覧を返す
func (srv *KojiService) GetKojies(
	ctx context.Context,
	req *grpcv1.GetKojiesRequest) (
	res *grpcv1.GetKojiesResponse,
	err error) {
	_ = req // 現状フィルター未対応

	// レスポンスを初期化
	res = grpcv1.GetKojiesResponse_builder{}.Build()

	grpcKojies := make(map[string]*grpcv1.Koji, srv.repo.Count())
	for _, grpcKoji := range srv.repo.GetAllAsMessage() {
		grpcKojies[grpcKoji.GetId()] = grpcKoji
	}

	res.SetKojies(grpcKojies)

	return
}

// GetKojiById は指定されたIDの工事データを返す
func (srv *KojiService) GetKoji(
	ctx context.Context,
	req *grpcv1.GetKojiRequest) (
	res *grpcv1.GetKojiResponse,
	err error) {

	// レスポンスを初期化
	res = grpcv1.GetKojiResponse_builder{}.Build()

	// リクエスト情報の取得
	id := req.GetTargetId()

	// 工事情報を取得
	koji, exist := srv.repo.Get(id)
	if !exist {
		err = connect.NewError(connect.CodeNotFound, errors.New("koji not found"))
		return
	}

	// Responseの更新
	res.SetKoji(koji.Message)

	return
}

func (srv *KojiService) UpdateKoji(
	ctx context.Context,
	req *grpcv1.UpdateKojiRequest) (
	res *grpcv1.UpdateKojiResponse,
	err error) {

	// ウォッチャーの一時停止
	if srv.watcher != nil {
		srv.watcher.Pause()
		defer srv.watcher.Resume()
	}

	// リクエスト情報の取得
	targetId := req.GetTargetId()
	sourceKojiMessage := req.GetSourceKoji()

	// 既存の工事情報を取得
	prevKoji, exist := srv.repo.Get(targetId)
	if !exist {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("koji not found"))
	}

	// 変更前の工事データのメッセージを保存
	prevMessageKoji, ok := proto.Clone(prevKoji.Message).(*grpcv1.Koji)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, errors.New("failed to assert koji message type"))
	}

	// 工事情報を更新
	err = srv.repo.Update(targetId, sourceKojiMessage)
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

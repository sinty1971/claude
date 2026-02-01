package services

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"

	grpcv1 "web-api/gen/grpc/v1"
	grpcv1connect "web-api/gen/grpc/v1/grpcv1connect"
	"web-api/internal/core"
	"web-api/internal/models"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

// CompanyService の実装
type CompanyService struct {
	// name はサービス名
	name string

	// cs は任意のgrpcサービスハンドラーへの参照
	cs *ContainerService

	// baseDirPath は会社一覧ディレクトリパス
	baseDirPath string

	// repo は会社情報リポジトリ（自動保存有効）
	repo *core.Repository[*models.Company]

	// watcher はファイルシステム監視オブジェクト
	watcher *core.Watcher

	// Embed the unimplemented handler for forward compatibility
	grpcv1connect.UnimplementedCompanyServiceHandler
}

func NewCompanyService(cs *ContainerService) *CompanyService {
	// パスをの取得と正規化
	baseDirPath, err := core.ResolveAbsPath(core.Config.CompanyBaseDirPath)
	if err != nil {
		panic(err)
	}

	return &CompanyService{
		name:        "CompanyService",
		baseDirPath: baseDirPath,
		repo:        core.NewRepository[*models.Company](true), // 自動保存有効
		cs:          cs,
	}
}

// Name はデータサービス名を返します
func (srv *CompanyService) Name() string {
	return srv.name
}

// GenerateHandler はサービスのハンドラを生成します
func (srv *CompanyService) GenerateHandler() (
	servicePath string,
	handler http.Handler,
	serviceName string) {

	// gRPC パスとハンドラの生成
	servicePath, handler = grpcv1connect.NewCompanyServiceHandler(srv)

	// サービス名の取得
	serviceName = grpcv1connect.CompanyServiceName

	return
}

// Start は CompanyListManager を初期化して開始します
func (srv *CompanyService) Start() error {
	// 全ての会社情報をキャッシュに取り込む
	srv.SyncAllToCache()

	// watcher の作成と初期化
	watcher, err := core.NewWatcher(srv.baseDirPath, core.Config.CompanyWatcherMaxDepth)
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

func (srv *CompanyService) Cleanup() {
	if srv.watcher != nil {
		srv.watcher.Close()
	}
}

// LoadAllCompanies は全ての会社情報をRepositoryに取り込みます
func (srv *CompanyService) SyncAllToCache() {
	// ファイルシステムから会社フォルダー一覧を取得
	entries, err := os.ReadDir(srv.baseDirPath)
	if err != nil {
		return
	}

	// Repositoryをクリア
	srv.repo.Clear()

	// 全てのCompanyインスタンスを作成
	for _, entry := range entries {
		// ディレクトリのみ処理
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(srv.baseDirPath, entry.Name())
		company, err := models.NewCompany(dirPath)
		cm, err := models.NewProtoModel[*models.Company]()
		if err != nil {
			continue
		}

		err = company.Load()
		if err != nil {
			log.Printf("マニフェストデータの読み込みに失敗しました 会社名 %s: %v", company.GetName(), err)
		}

		// Repositoryに追加（初期ロード時は自動保存しない）
		srv.repo.SetAutoSave(false)
		if err := srv.repo.Set(company.GetId(), company); err != nil {
			log.Printf("リポジトリへの追加に失敗しました: %v", err)
		}
	}

	// 自動保存を有効化
	srv.repo.SetAutoSave(true)
}

// Update は指定 targetId のRepository情報を新しい会社情報で更新します（自動保存）
//
//	targetId: 更新対象会社Id
//	source: 新しい会社情報、nilの場合は更新対象会社自身の更新を行います
func (srv *CompanyService) Update(targetId string, source *models.Company) error {
	// Repository経由で更新（自動的にSave()が呼ばれる）
	return srv.repo.Update(targetId, func(target *models.Company) error {
		return target.Update(source)
	})
}

// watchFileSystemEvents はファイルシステム監視イベントを処理します
func (srv *CompanyService) watchFileSystemEvents() {
	for {
		select {
		case event, ok := <-srv.watcher.Events():
			if !ok {
				return
			}

			// eventから会社Idの取得
			dirName := filepath.Base(filepath.Dir(event.Name))
			id := core.GenerateIdFromString(dirName)

			// 会社情報の存在チェック
			_, exist := srv.repo.Get(id)
			if !exist {
				srv.SyncAllToCache()
			} else {
				err := srv.Update(id, nil)
				if err != nil {
					log.Printf("CompanyService: Failed to update company cache map for known company: %s, Error: %v", id, err)
				}
			}

		case err, ok := <-srv.watcher.Errors():
			if !ok {
				return
			}
			log.Printf("CompanyService: File system watcher error: %v", err)
		}
	}
}

// GetCompanies は管理されている会社情報の一覧を取得します
// gRPCサービスの実装です
func (srv *CompanyService) GetCompanies(
	// args
	_ context.Context,
	req *grpcv1.GetCompaniesRequest) (

	// returns
	res *grpcv1.GetCompaniesResponse,
	err error) {

	// レスポンスを初期化
	res = grpcv1.GetCompaniesResponse_builder{}.Build()

	// 必要に応じてキャッシュを更新
	if req.GetForceReload() {
		srv.SyncAllToCache()
	}

	// 会社データモデルを作成
	allCompanies := srv.repo.GetAll()
	grpcv1Companies := make(map[string]*grpcv1.Company, len(allCompanies))
	for _, v := range allCompanies {
		if v != nil && v.Company != nil {
			grpcv1Companies[v.GetId()] = v.Company
		}
	}

	// Responseの更新とリターン
	res.SetCompanies(grpcv1Companies)
	return res, nil
}

// GetCompany は会社IDから会社情報を取得します
// gRPCサービスの実装です
func (srv *CompanyService) GetCompany(
	_ context.Context,
	req *grpcv1.GetCompanyRequest) (
	res *grpcv1.GetCompanyResponse,
	err error) {

	// レスポンスを初期化
	res = grpcv1.GetCompanyResponse_builder{}.Build()

	// Idの取得
	id := req.GetTargetId()

	// 会社情報を取得
	company, exist := srv.repo.Get(id)
	if !exist {
		err = connect.NewError(connect.CodeNotFound, errors.New("company not found"))
		return
	}

	// Responseの更新
	res.SetCompany(company.Company)

	return
}

// UpdateCompany は会社情報を更新します
// gRPCサービスの実装です
// 既存の Id の会社情報を更新します。そのため Id の変更の可能性があります。
// また、フォルダーの移動も発生する可能性があります。
func (srv *CompanyService) UpdateCompany(
	// args
	_ context.Context,
	req *grpcv1.UpdateCompanyRequest) (

	// returns
	res *grpcv1.UpdateCompanyResponse,
	err error) {

	// リクエスト情報の取得
	targetId := req.GetTargetId()
	target, exist := srv.repo.Get(targetId)
	if !exist {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("指定された target_id の会社データが存在しません"))
	}

	// prevMessageCompany の作成
	prevMessageCompany := proto.Clone(target.Company).(*grpcv1.Company)

	// source(proto.Message) から Company モデルを作成
	source, err := models.NewCompanyFromMessage(req.GetSourceCompany())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if srv.watcher != nil {
		srv.watcher.Pause()
		defer srv.watcher.Resume()
	}

	err = srv.Update(targetId, source)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Responseの作成
	res = grpcv1.UpdateCompanyResponse_builder{}.Build()
	res.SetPrevCompany(prevMessageCompany)

	return res, nil
}

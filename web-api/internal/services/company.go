package services

import (
	"context"
	"errors"
	"log"
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
	// CS は任意のgrpcサービスハンドラーへの参照
	CS *ContainerService

	// DirPath は会社一覧フォルダーパス
	DirPath string

	// cache は会社情報キャッシュマップ
	cache map[string]*models.Company

	// watcher はファイルシステム監視オブジェクト
	watcher *core.Watcher

	// Embed the unimplemented handler for forward compatibility
	grpcv1connect.UnimplementedCompanyServiceHandler
}

// Name はデータサービス名を返します
func (srv *CompanyService) Name() string {
	return "CompanyService"
}

// Start は CompanyListManager を初期化して開始します
func (srv *CompanyService) Start(cs *ContainerService) error {

	// 既存インスタンスに値をセット（再代入しないこと）
	srv.CS = cs

	// パスをの取得と正規化
	dirPath, err := core.NormalizeAbsPath(core.Config.CompanyServiceDirPath)
	if err != nil {
		return err
	}
	srv.DirPath = dirPath

	// 全ての会社情報をデータベースに取り込む
	err = srv.SyncAllToCache()
	if err != nil {
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

func (srv *CompanyService) Cleanup() {
	if srv.watcher != nil {
		srv.watcher.Close()
	}
}

// consumeWatcherEvents はファイルシステム監視イベントを処理します
func (srv *CompanyService) consumeWatcherEvents() {
	for {
		select {
		case event, ok := <-srv.watcher.Events():
			if !ok {
				return
			}
			log.Printf("CompanyService: File system event: %s", event)

			// データベースへの同期
			if err := srv.SyncAllToCache(); err != nil {
				log.Printf("CompanyService: Failed to update company cache map: %v", err)
			}

		case err, ok := <-srv.watcher.Errors():
			if !ok {
				return
			}
			log.Printf("CompanyService: File system watcher error: %v", err)
		}
	}
}

// LoadAllCompanies は全ての会社情報をキャッシュに取り込みます
func (srv *CompanyService) SyncAllToCache() error {
	// ファイルシステムから会社フォルダー一覧を取得
	entries, err := os.ReadDir(srv.DirPath)
	if err != nil {
		return err
	}

	// データの初期化
	srv.cache = make(map[string]*models.Company, len(entries))

	// 全てのCompanyインスタンスを作成
	for _, entry := range entries {
		// Companyインスタンスの作成と初期化
		dirPath := filepath.Join(srv.DirPath, entry.Name())
		company, err := models.NewCompany(dirPath)
		if err == nil {
			srv.cache[company.GetId()] = company
		}
	}

	// Manifest情報の読み込み
	for _, company := range srv.cache {
		err := company.Manifest.Load()
		if err != nil {
			log.Printf("マニフェストの永続化情報の読み込みに失敗しました 会社略称 %s: %v", company.GetShortName(), err)
		}
	}

	return nil
}

// Update は指定 targetId のキャッシュ情報を新しい会社情報で更新します
//
//	targetId: 更新対象会社Id
//	source: 新しい会社情報
func (srv *CompanyService) Update(targetId string, src *models.Company) error {
	// 引数チェック
	if src == nil {
		return errors.New("更新情報 src の値が nil です")
	}

	// targetId からCompanyデータを取得
	target, exist := srv.cache[targetId]
	if !exist {
		return errors.New("更新対象の会社情報が存在しません")
	}

	// 会社情報の更新
	err := target.Update(src)
	if err != nil {
		return err
	}

	// キャッシュ情報の更新
	srv.cache[target.GetId()] = target

	return nil
}

// GetCompanies は管理されている会社情報の一覧を取得します
// gRPCサービスの実装です
func (srv *CompanyService) GetCompanies(
	// args
	ctx context.Context,
	req *grpcv1.GetCompaniesRequest) (

	// returns
	res *grpcv1.GetCompaniesResponse,
	err error) {

	// レスポンスを初期化
	res = grpcv1.GetCompaniesResponse_builder{}.Build()

	// 必要に応じてキャッシュを更新
	if req.GetForceReload() {
		if err := srv.SyncAllToCache(); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// 会社データモデルを作成
	grpcv1Companies := make(map[string]*grpcv1.Company, len(srv.cache))
	for _, v := range srv.cache {
		grpcv1Companies[v.Company.GetId()] = v.Company
	}

	// Responseの更新とリターン
	res.SetCompanies(grpcv1Companies)
	return res, nil
}

// GetCompany は会社IDから会社情報を取得します
// gRPCサービスの実装です
func (srv *CompanyService) GetCompany(
	ctx context.Context,
	req *grpcv1.GetCompanyRequest) (
	res *grpcv1.GetCompanyResponse,
	err error) {

	// レスポンスを初期化
	res = grpcv1.GetCompanyResponse_builder{}.Build()

	// Idの取得
	id := req.GetTargetId()

	// 会社情報を取得
	company, exist := srv.cache[id]
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
	ctx context.Context,
	req *grpcv1.UpdateCompanyRequest) (

	// returns
	res *grpcv1.UpdateCompanyResponse,
	err error) {

	// リクエスト情報の取得
	targetId := req.GetTargetId()
	target, exist := srv.cache[targetId]
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

	err = srv.Update(targetId, source)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Responseの作成
	res = grpcv1.UpdateCompanyResponse_builder{}.Build()
	res.SetPrevCompany(prevMessageCompany)

	return res, nil
}

// GetCompanyCategories は業種カテゴリーの一覧を取得します
func (srv *CompanyService) GetCompanyCategories(
	_ context.Context, _ *grpcv1.GetCompanyCategoriesRequest) (
	*grpcv1.GetCompanyCategoriesResponse, error) {

	// レスポンスを初期化
	res := grpcv1.GetCompanyCategoriesResponse_builder{}.Build()

	categories := make([]*grpcv1.CompanyCategory, 0, len(models.CompanyCategoryMap))
	for idx, label := range models.CompanyCategoryMap {
		categories = append(categories, grpcv1.CompanyCategory_builder{
			Index: int32(idx),
			Label: label,
		}.Build())
	}

	res.SetCategories(categories)

	return res, nil
}

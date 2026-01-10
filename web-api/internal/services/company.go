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
)

// CompanyService は会社情報管理サービスを提供します
type CompanyService struct {
	// dirPath は会社一覧ディレクトリパス
	dirPath string

	// watcher はファイルシステム監視オブジェクト
	watcher *core.Watcher

	// container は各サービスをまとめるコンテナへの参照
	container *ServiceContainer

	// Embed the unimplemented handler for forward compatibility
	grpcv1connect.UnimplementedCompanyServiceHandler
}

// Name はデータサービス名を返します
func (srv *CompanyService) Name() string {
	return "CompanyService"
}

// Start は CompanyService を初期化して開始します
func (srv *CompanyService) Start(container *ServiceContainer) error {

	// 既存インスタンスに値をセット（再代入しないこと）
	srv.container = container

	// パスをの取得と正規化
	dirPath, err := core.NormalizeAbsPath(core.Config.CompanyServiceFolder)
	if err != nil {
		return err
	}
	srv.dirPath = dirPath

	// 全ての会社情報をデータベースに取り込む
	if err = srv.LoadAllCompanies(); err != nil {
		return err
	}

	// オプションが存在する場合は数値に変換
	watcher, err := core.NewWatcher()
	if err != nil {
		return err
	}
	srv.watcher = watcher

	if err = srv.watcher.Start(
		srv.dirPath,
		core.Config.CompanyWatcherMaxDepth); err != nil {
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

			// 会社キャッシュの更新
			if err := srv.LoadAllCompanies(); err != nil {
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

// LoadAllCompanies は全ての会社情報をデータベースに取り込みます
func (srv *CompanyService) LoadAllCompanies() error {
	// ファイルシステムから会社フォルダー一覧を取得
	entries, err := os.ReadDir(srv.dirPath)
	if err != nil {
		return err
	}

	// データの初期化
	companies := make(map[string]*models.Company, len(entries))

	// 全てのCompanyインスタンスを作成
	for _, entry := range entries {
		// Companyインスタンスの作成と初期化
		company := models.NewCompany()
		if err := company.ParseFrom(srv.dirPath, entry.Name()); err == nil {
			companies[company.GetId()] = company
		}
	}

	// Manifest データの読み込み
	for _, company := range companies {
		if err := company.Load(); err != nil {
			log.Printf("Failed to load manifest data for company ShortName %s: %v", company.GetShortName(), err)
		}
	}

	// データベースへインポート
	if err := srv.importToDB(); err != nil {
		return err
	}
	return nil
}

// syncFromDirectory はディレクトリから会社情報を読み込み、
// データベースに同期します
func (srv *CompanyService) syncFromDirectory() error {
	// ディレクトリエントリの取得
	entries, err := os.ReadDir(srv.dirPath)
	if err != nil {
		return err
	}

	// 会社情報の生成と加工
	companies := make(map[string]*models.Company, len(entries))
	for _, entry := range entries {
		company := models.NewCompany()
		if err := company.ParseFrom(srv.dirPath, entry.Name()); err == nil {
			companies[company.GetId()] = company
		}
	}

	// Manifest データの読み込み
	for _, company := range companies {
		if err := company.Load(); err != nil {
			log.Printf("Failed to load manifest: %v", err)
		}
	}

	// TODO: データベースへインポート
	return nil
}

// UpdateNewCompany は指定 id のキャッシュ情報を新しい会社情報で更新します
// prevId: 更新対象の会社ID、存在しない場合は追加
// newCompany: 更新後の会社情報
func (srv *CompanyService) UpdateNewCompany(prevId string, newCompany *models.Company) (*models.Company, error) {
	// Idから更新前の会社情報を取得
	prevCompany, exist := srv.companies[prevId]

	// キャッシュから削除
	if exist {
		delete(srv.companies, prevId)
	}

	// 新しい会社情報の管理フォルダー名を生成
	newTarget := models.GenerateCompanyPathistFolder(
		filepath.Dir(newCompany.GetPathistFolder()),
		newCompany.GetCategoryIndex(),
		newCompany.GetShortName())

	// 管理フォルダーの変更がある場合はフォルダー移動を実施
	if exist && prevCompany.GetPathistFolder() != newCompany.GetPathistFolder() {
		// 管理フォルダーの変更がある場合はフォルダー移動を実施
		prevTarget := prevCompany.GetPathistFolder()
		if err := os.Rename(prevTarget, newTarget); err != nil {
			return nil, err
		}
	}

	// 既存の情報を更新
	srv.companies[newCompany.GetId()] = newCompany

	// persist情報の書き込み
	if err := newCompany.Persist.SavePersists(); err != nil {
		log.Printf("Failed to save persist info for company ShortName %s: %v", newCompany.GetShortName(), err)
	}

	return prevCompany, nil
}

// GetCompanies は管理されている会社情報の一覧を取得します
// gRPCサービスの実装です
func (srv *CompanyService) GetCompanies(
	ctx context.Context, req *grpcv1.GetCompaniesRequest) (
	*grpcv1.GetCompaniesResponse, error) {

	// レスポンスを初期化
	res := grpcv1.GetCompaniesResponse_builder{}.Build()

	// 必要に応じてキャッシュを更新
	if req.GetRefresh() {
		if err := srv.UpdateCompanies(); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// 会社データモデルを作成
	grpcv1Companies := make(map[string]*grpcv1.Company, len(srv.companies))
	for _, v := range srv.companies {
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
	id := req.GetId()

	// 会社情報を取得
	company, exist := srv.companies[id]
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
// Company.Id 更新対象の会社Id
func (srv *CompanyService) UpdateCompany(
	_ context.Context, req *grpcv1.UpdateCompanyRequest) (
	*grpcv1.UpdateCompanyResponse, error) {

	// リクエスト情報の取得
	prevId := req.GetPrevId()
	newCompany := &models.Company{Company: req.GetNewCompany()}

	prevCompany, err := srv.UpdateNewCompany(prevId, newCompany)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("updated company is nil"))
	}

	// Responseの作成
	res := grpcv1.UpdateCompanyResponse_builder{}.Build()
	res.SetPrevCompany(prevCompany.Company)

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

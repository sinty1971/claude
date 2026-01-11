package ctrl

import (
	"context"
	"database/sql"
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

// CompanyService の実装
type CompanyService struct {
	// CS は任意のgrpcサービスハンドラーへの参照
	CS *ContainerService

	// DirPath は会社一覧フォルダーパス
	DirPath string

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
	dirPath, err := core.NormalizeAbsPath(core.Config.CompanyServiceFolder)
	if err != nil {
		return err
	}
	srv.DirPath = dirPath

	// 全ての会社情報をデータベースに取り込む
	err = srv.SyncAllToDb()
	if err != nil {
		return err
	}

	// オプションが存在する場合は数値に変換
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
			if err := srv.SyncAllToDb(); err != nil {
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
func (srv *CompanyService) SyncAllToDb() error {
	// ファイルシステムから会社フォルダー一覧を取得
	entries, err := os.ReadDir(srv.DirPath)
	if err != nil {
		return err
	}

	// データの初期化
	cache := make(map[string]*models.Company, len(entries))

	// 全てのCompanyインスタンスを作成
	for _, entry := range entries {
		// Companyインスタンスの作成と初期化
		dirPath := filepath.Join(srv.DirPath, entry.Name())
		company, err := models.NewCompany(dirPath)
		if err == nil {
			cache[company.GetId()] = company
		}
	}

	// Manifest情報の読み込み
	for _, company := range cache {
		err := company.Manifest.Load()
		if err != nil {
			log.Printf("マニフェストの永続化情報の読み込みに失敗しました 会社略称 %s: %v", company.GetShortName(), err)
		}
	}

	// ContainerService の確認
	if srv.CS == nil {
		return errors.New("ContainerService の値が nil です")
	}

	// DBハンドラーの確認
	if srv.CS.DB() == nil {
		return errors.New("データベースハンドラの値が nil です")
	}

	// トランザクション開始
	tx, err := srv.CS.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 既存データの削除
	if _, err := tx.Exec(`DELETE FROM companies`); err != nil {
		return err
	}

	insertSQL, err := core.BuildInsertSQLFromDefaultMigrations("companies")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, company := range cache {
		_, err := stmt.Exec(
			company.GetId(),
			company.GetMfLongName(),
			company.GetMfPostalCode(),
			company.GetMfAddress(),
			company.GetMfTel(),
			company.GetMfFax(),
			company.GetMfEmail(),
			company.GetMfWebsite(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateNewCompany は指定 id のキャッシュ情報を新しい会社情報で更新します
// prevId: 更新対象の会社ID、存在しない場合は追加
// newCompany: 更新後の会社情報
func (srv *CompanyService) SyncToDB(ids []string) error {
	// ファイルシステムから会社フォルダー一覧を取得
	entries, err := os.ReadDir(srv.DirPath)
	if err != nil {
		return err
	}

	// データの初期化
	cache := make(map[string]*models.Company, len(entries))

	// 全てのCompanyインスタンスを作成
	for _, entry := range entries {
		// Companyインスタンスの作成と初期化
		dirPath := filepath.Join(srv.DirPath, entry.Name())
		company, err := models.NewCompany(dirPath)
		if err == nil {
			cache[company.GetId()] = company
		}
	}

	// Manifest情報の読み込み
	for _, company := range cache {
		err := company.Manifest.Load()
		if err != nil {
			log.Printf("マニフェストの永続化情報の読み込みに失敗しました 会社略称 %s: %v", company.GetShortName(), err)
		}
	}

	// ContainerService の確認
	if srv.CS == nil {
		return errors.New("ContainerService の値が nil です")
	}

	// DBハンドラーの確認
	if srv.CS.DB() == nil {
		return errors.New("データベースハンドラの値が nil です")
	}

	// トランザクション開始
	tx, err := srv.CS.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 既存データの削除
	if _, err := tx.Exec(`DELETE FROM companies`); err != nil {
		return err
	}

	insertSQL, err := core.BuildInsertSQLFromDefaultMigrations("companies")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, company := range cache {
		_, err := stmt.Exec(
			company.GetId(),
			company.GetMfLongName(),
			company.GetMfPostalCode(),
			company.GetMfAddress(),
			company.GetMfTel(),
			company.GetMfFax(),
			company.GetMfEmail(),
			company.GetMfWebsite(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetCompanyByIdFromDB はデータベースから会社情報を取得します
func (srv *CompanyService) GetCompanyByIdFromId(id string) (*grpcv1.Company, error) {
	if id == "" {
		return nil, errors.New("id が空です")
	}

	if srv.CS == nil {
		return nil, errors.New("ContainerService の値が nil です")
	}
	if srv.CS.DB() == nil {
		return nil, errors.New("データベースハンドラの値が nil です")
	}

	selectSQL, err := core.BuildSelectByIDSQLFromDefaultMigrations("companies")
	if err != nil {
		return nil, err
	}

	row := srv.CS.DB().QueryRow(selectSQL, id)

	var (
		companyID  string
		longName   string
		postalCode string
		address    string
		tel        string
		fax        string
		email      string
		website    string
	)

	if err := row.Scan(
		&companyID,
		&longName,
		&postalCode,
		&address,
		&tel,
		&fax,
		&email,
		&website,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("company not found")
		}
		return nil, err
	}

	return grpcv1.Company_builder{
		Id:           companyID,
		MfLongName:   longName,
		MfPostalCode: postalCode,
		MfAddress:    address,
		MfTel:        tel,
		MfFax:        fax,
		MfEmail:      email,
		MfWebsite:    website,
	}.Build(), nil
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

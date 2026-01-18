package services

import (
	"context"
	"net/http"
	grpcv1 "web-api/gen/grpc/v1"
	grpcv1connect "web-api/gen/grpc/v1/grpcv1connect"
	"web-api/internal/core"
)

// defaultCompanyCategoryMap 業種カテゴリーのラベルマップです
// 将来的にはyamlファイルから読み込む予定
var defaultCompanyCategoryMap = map[int32]string{
	0: "自社組合",
	1: "下請会社",
	2: "築炉会社",
	3: "一人親方",
	4: "元請会社",
	5: "リース会社",
	6: "販売会社",
	7: "販売会社２",
	8: "求人会社",
	9: "一般会社",
}

// CompanyCategoryService は会社業種カテゴリーに関する gRPC サービスハンドラです
type CompanyCategoryService struct {
	// name はサービス名
	name string

	//
	cs *ContainerService

	// baseDirPath はカテゴリーマニフェスト保持ディレクトリパス
	baseDirPath string

	// cache は会社情報キャッシュマップ
	cache map[string]*grpcv1.CompanyCategory

	// watcher はファイルシステム監視オブジェクト
	watcher *core.Watcher

	// Embed the unimplemented handler for forward compatibility
	grpcv1connect.UnimplementedCompanyCategoryServiceHandler
}

// NewCompanyCategoryService は新しい CompanyCategoryService インスタンスを作成します
func NewCompanyCategoryService(cs *ContainerService) *CompanyCategoryService {
	return &CompanyCategoryService{
		name: "CompanyCategoryService",
		cs:   cs,
	}
}

// NewCompanyCategory インスタンス作成と初期化を行います
func NewCompanyCategory(idx int32, name string) (cc *grpcv1.CompanyCategory) {
	cc = grpcv1.CompanyCategory_builder{}.Build()
	cc.SetIndex(idx)
	cc.SetName(name)

	return
}

// GenerateHandler はサービスのハンドラを生成します
func (srv *CompanyCategoryService) GenerateHandler() (
	servicePath string, handler http.Handler, serviceName string) {

	// gRPC パスとハンドラの生成
	servicePath, handler = grpcv1connect.NewCompanyCategoryServiceHandler(srv)

	// サービス名の取得
	serviceName = grpcv1connect.CompanyCategoryServiceName

	return
}

// Name はデータサービス名を返します
func (srv *CompanyCategoryService) Name() string {
	return srv.name
}

// Start は CompanyListManager を初期化して開始します
func (srv *CompanyCategoryService) Start() error {
	// 全ての会社情報をキャッシュに取り込む
	err := srv.SyncAllToCache()
	if err != nil {
		return err
	}

	core.CompanyCategoryMin = 0
	catMat := int32(0)
	for _, cat := range srv.cache {
		if int(cat.GetIndex()) > int(catMat) {
			catMat = cat.GetIndex()
		}
	}
	core.CompanyCategoryMax = catMat

	// CompanyCategoryService は静的データのため watcher は不要
	// watcher の作成をスキップ

	return nil
}

// Cleanup はサービスのクリーンアップ処理を実行します
func (srv *CompanyCategoryService) Cleanup() {
	if srv.watcher != nil {
		srv.watcher.Close()
	}
}

func (srv *CompanyCategoryService) SyncAllToCache() error {
	// キャッシュマップを初期化
	srv.cache = make(map[string]*grpcv1.CompanyCategory, len(defaultCompanyCategoryMap))

	// 全てのCompanyCategoryインスタンスを作成
	for idx, name := range defaultCompanyCategoryMap {
		category := NewCompanyCategory(idx, name)
		srv.cache[name] = category
		core.CompanyCategoryCacheMap[name] = category
	}
	return nil
}

func (srv *CompanyCategoryService) WatchEvents() {
	// 現在は監視イベントの処理は不要
}

// GetCompanyCategories は業種カテゴリーの一覧を取得します
func (srv *CompanyCategoryService) GetCompanyCategories(
	// args
	_ context.Context,
	_ *grpcv1.GetCompanyCategoriesRequest) (
	// returns
	res *grpcv1.GetCompanyCategoriesResponse,
	err error) {
	// CompanyCategory 配列の作成
	categories := make([]*grpcv1.CompanyCategory, 0, len(srv.cache))
	for _, category := range srv.cache {
		categories = append(categories, category)
	}

	// レスポンスを初期化と設定
	res = grpcv1.GetCompanyCategoriesResponse_builder{}.Build()
	res.SetCategories(categories)

	return
}

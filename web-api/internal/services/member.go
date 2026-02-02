package services

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	grpcv1 "web-api/gen/grpc/v1"
	grpcv1connect "web-api/gen/grpc/v1/grpcv1connect"
	"web-api/internal/core"
	"web-api/internal/models"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
)

// MemberService は Member に関する gRPC サービスハンドラです
type MemberService struct {
	// name はサービス名
	name string

	// cs は任意のgrpcサービスハンドラーへの参照
	cs *ContainerService

	// 作業員一覧ディレクトリパス
	baseDirPath string

	// repo はMember情報リポジトリ（自動保存有効）
	repo *core.Repository[*models.Member]

	// watcher はファイルシステム監視オブジェクト
	watcher *core.Watcher

	// Embed the unimplemented handler for forward compatibility
	grpcv1connect.UnimplementedMemberServiceHandler
}

// NewMemberService は新しい MemberService インスタンスを作成します
func NewMemberService(cs *ContainerService) *MemberService {
	// パスをの取得と正規化
	baseDirPath, err := core.ResolveAbsPath(core.Config.MemberBaseDirPath)
	if err != nil {
		panic(err)
	}

	return &MemberService{
		name:        "MemberService",
		cs:          cs,
		baseDirPath: baseDirPath,
		repo:        core.NewRepository[*models.Member](true), // 自動保存有効
	}
}

// Name はサービス名を返します
func (srv *MemberService) Name() string {
	return srv.name
}

// GenerateHandler はサービスのハンドラを生成します
func (srv *MemberService) GenerateHandler() (
	servicePath string, handler http.Handler, serviceName string) {

	// gRPC パスとハンドラの生成
	servicePath, handler = grpcv1connect.NewMemberServiceHandler(srv)

	// サービス名の所得
	serviceName = grpcv1connect.MemberServiceName

	return
}

// Start はサービスを開始します（現在はプレースホルダー）
func (srv *MemberService) Start() error {
	// 初期データのロード処理などをここに実装予定
	err := srv.SyncAllToCache()
	if err != nil {
		return err
	}

	return nil
}

// Cleanup はサービスのクリーンアップを行います
func (srv *MemberService) Cleanup() {
	// リソースのクリーンアップ処理をここに実装予定
}

func (srv *MemberService) SyncAllToCache() error {
	// Repositoryをクリア
	srv.repo.Clear()

	// ターゲットディレクトリの抽出
	requests := srv.extractTargetRequests()

	// 全てのMemberインスタンスを作成
	for _, request := range requests {

		member, err := models.NewPersistModelMember(request.GetDirPath())
		if err != nil {
			continue
		}

		err = member.Load()
		if err != nil {
			mes := member.Message.(*grpcv1.Member)
			log.Printf("マニフェストデータの読み込みに失敗しました 作業員名 %s: %v", mes.GetName(), err)
		}

		// Repositoryに追加（初期ロード時は自動保存しない）
		srv.repo.SetAutoSave(false)
		if err := srv.repo.Set(*member); err != nil {
			log.Printf("リポジトリへの追加に失敗しました: %v", err)
		}
	}

	// 自動保存を有効化
	srv.repo.SetAutoSave(true)

	return nil
}

// 対象ディレクトリの抽出
func (srv *MemberService) extractTargetRequests() (requests []*grpcv1.Member) {
	// 対象ディレクトリ
	// .../'1 会社'/'3 個人名' の形式
	// .../'1 会社'/[会社名]/'社員'/[メンバー名] の形式
	// .../'1 会社'/[会社名]/'社員'/'@退職者'/[メンバー名] の形式
	entries, err := os.ReadDir(srv.baseDirPath)
	if err != nil {
		return
	}

	// 十分な容量を確保
	requests = make([]*grpcv1.Member, 0, len(entries)*100)

	for _, entry := range entries {
		// ディレクトリのみ処理
		if !entry.IsDir() {
			continue
		}

		// CompanyCategoryの取得
		companyDirname := entry.Name()
		cat, err := core.ParseCompanyCategoryFromDirName(companyDirname)
		if err != nil {
			continue
		}

		// request
		request := grpcv1.Member_builder{}.Build()

		// 一人親方ディレクトリである場合の処理
		if core.ContainCompanyCategoryName(cat, "一人親方") {
			dirPath := filepath.Join(srv.baseDirPath, entry.Name())
			request.SetDirPath(dirPath)
			requests = append(requests, request)
			continue
		}

		// 会社カテゴリーが自社組合の場合の処理
		if core.ContainCompanyCategoryName(cat, "自社組合") ||
			core.ContainCompanyCategoryName(cat, "下請会社") ||
			core.ContainCompanyCategoryName(cat, "築炉会社") {
			activeDirPath := filepath.Join(srv.baseDirPath, entry.Name(), "社員")
			activeEntries, err := os.ReadDir(activeDirPath)
			if err != nil {
				continue
			}
			for _, activeEntry := range activeEntries {
				if !activeEntry.IsDir() {
					continue
				}
				if strings.HasPrefix(activeEntry.Name(), "@") {
					continue
				}
				dirPath := filepath.Join(activeDirPath, activeEntry.Name())
				request.SetDirPath(dirPath)
				requests = append(requests, request)
			}

			deactiveDirPath := filepath.Join(srv.baseDirPath, entry.Name(), "社員", "@退職者")
			deactiveEntries, err := os.ReadDir(deactiveDirPath)
			if err != nil {
				continue
			}
			for _, deactiveEntry := range deactiveEntries {
				if !deactiveEntry.IsDir() {
					continue
				}
				dirPath := filepath.Join(deactiveDirPath, deactiveEntry.Name())
				request.SetDirPath(dirPath)
				requests = append(requests, request)
			}
		}
	}
	return
}

// GetMembers は全ての Member 情報を返します
func (srv *MemberService) GetMembers(
	ctx context.Context,
	req *grpcv1.GetMembersRequest) (
	res *grpcv1.GetMembersResponse,
	err error) {
	_ = req // 現状フィルター未対応

	// レスポンスの初期化
	res = grpcv1.GetMembersResponse_builder{}.Build()

	// Repositoryから全てのMemberを取得
	grpcMembers := make(map[string]*grpcv1.Member, srv.repo.Count())
	for _, mes := range srv.repo.GetAllAsMessage() {
		grpcMember, ok := mes.(*grpcv1.Member)
		if !ok {
			continue
		}
		grpcMembers[grpcMember.GetId()] = grpcMember
	}

	// レスポンスの設定とリターン
	res.SetMembers(grpcMembers)
	return
}

// GetMember は指定された ID の Member 情報を返します
func (srv *MemberService) GetMember(
	ctx context.Context,
	req *grpcv1.GetMemberRequest) (
	res *grpcv1.GetMemberResponse,
	err error) {

	// レスポンスの初期化
	res = grpcv1.GetMemberResponse_builder{}.Build()

	// targetIdの取得
	targetId := req.GetTargetId()

	member, exists := srv.repo.Get(targetId)
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	grpcMember, ok := member.Message.(*grpcv1.Member)
	if !ok {
		err = connect.NewError(connect.CodeInternal, errors.New("failed to assert member message type"))
		return
	}

	// レスポンスの設定とリターン
	res.SetMember(grpcMember)

	return
}

// UpdateMember は Member 情報を更新します
func (srv *MemberService) UpdateMember(
	ctx context.Context,
	req *grpcv1.UpdateMemberRequest) (
	res *grpcv1.UpdateMemberResponse,
	err error) {

	// レスポンスの初期化
	res = grpcv1.UpdateMemberResponse_builder{}.Build()

	// リクエスト情報の取得
	targetId := req.GetTargetId()
	sourceGrpcMember := req.GetSourceMember()

	// 既存の Member 情報を取得
	prevMember, exists := srv.repo.Get(targetId)
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	// prevGrpcMember の作成
	prevGrpcMember := proto.Clone(prevMember.Message).(*grpcv1.Member)

	err = srv.repo.Update(targetId, sourceGrpcMember)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// レスポンスの設定とリターン
	res.SetPrevMember(prevGrpcMember)

	return
}

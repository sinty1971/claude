package services

import (
	"context"
	"log"
	"os"
	"path/filepath"

	grpcv1 "web-api/gen/grpc/v1"
	grpcv1connect "web-api/gen/grpc/v1/grpcv1connect"
	"web-api/internal/core"
	"web-api/internal/models"

	"connectrpc.com/connect"
)

// MemberService は Member に関する gRPC サービスハンドラです
type MemberService struct {
	// name はサービス名
	name string

	// cs は任意のgrpcサービスハンドラーへの参照
	cs *ContainerService

	// 作業員一覧ディレクトリパス
	baseDirPath string

	// cache は Member 情報キャッシュマップ（ID をキーとする）
	cache map[string]*models.Member

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
		cache:       make(map[string]*models.Member),
	}
}

// Name はサービス名を返します
func (srv *MemberService) Name() string {
	return srv.name
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
	// ファイルシステムから作業員フォルダー一覧を取得
	entries, err := os.ReadDir(srv.baseDirPath)
	if err != nil {
		return err
	}

	// キャッシュマップを初期化
	srv.cache = make(map[string]*models.Member, len(entries))

	// 全てのMemberインスタンスを作成
	for _, entry := range entries {
		// ディレクトリのみ処理
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(srv.baseDirPath, entry.Name())
		member, err := models.NewMember(dirPath)
		if err != nil {
			continue
		}

		err = member.LoadManifest()
		if err != nil {
			log.Printf("マニフェストデータの読み込みに失敗しました 作業員名 %s: %v", member.GetName(), err)
		}

		srv.cache[member.GetId()] = member
	}

	return nil
}

// GetMembers は全ての Member 情報を返します
func (srv *MemberService) GetMembers(
	ctx context.Context,
	req *connect.Request[grpcv1.GetMembersRequest],
) (*connect.Response[grpcv1.GetMembersResponse], error) {
	// キャッシュから全ての Member を取得
	members := make(map[string]*grpcv1.Member, len(srv.cache))
	for id, member := range srv.cache {
		members[id] = member.Member
	}

	response := grpcv1.GetMembersResponse_builder{}.Build()
	response.SetMembers(members)

	return connect.NewResponse(response), nil
}

// GetMember は指定された ID の Member 情報を返します
func (srv *MemberService) GetMember(
	ctx context.Context,
	req *connect.Request[grpcv1.GetMemberRequest],
) (*connect.Response[grpcv1.GetMemberResponse], error) {
	id := req.Msg.GetTargetId()

	member, exists := srv.cache[id]
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	response := grpcv1.GetMemberResponse_builder{}.Build()
	response.SetMember(member.Member)

	return connect.NewResponse(response), nil
}

// UpdateMember は Member 情報を更新します
func (srv *MemberService) UpdateMember(
	ctx context.Context,
	req *connect.Request[grpcv1.UpdateMemberRequest],
) (*connect.Response[grpcv1.UpdateMemberResponse], error) {
	targetId := req.Msg.GetTargetId()
	sourceMember := req.Msg.GetSourceMember()

	if sourceMember == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, nil)
	}

	prevMember, exists := srv.cache[targetId]
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	// 新しい Member インスタンスを作成
	member, err := models.NewMemberFromMessage(sourceMember)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// キャッシュに保存
	srv.cache[targetId] = member

	// TODO: Manifest ファイルへの永続化処理を実装

	response := grpcv1.UpdateMemberResponse_builder{}.Build()
	response.SetPrevMember(prevMember.Member)

	return connect.NewResponse(response), nil
}

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
	// ターゲットディレクトリの抽出
	targetDirs := srv.extractTargetDirPaths()

	// キャッシュマップを初期化
	srv.cache = make(map[string]*models.Member, len(targetDirs))

	// 全てのMemberインスタンスを作成
	for _, dirPath := range targetDirs {

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

// Update はメンバー情報を更新します
func (srv *MemberService) Update(targetId string, source *models.Member) error {
	// targetId から Member データを取得
	target, exist := srv.cache[targetId]
	if !exist {
		return errors.New("更新対象のメンバー情報が存在しません")
	}

	// メンバー情報の更新
	err := target.Update(source)
	if err != nil {
		return err
	}

	// キャッシュ情報の更新
	srv.cache[target.GetId()] = target

	return nil
}

// 対象ディレクトリの抽出
func (srv *MemberService) extractTargetDirPaths() (targetDirs []string) {
	// 対象ディレクトリ
	// .../'1 会社'/'3 個人名' の形式
	// .../'1 会社'/[会社名]/'社員'/[メンバー名] の形式
	// .../'1 会社'/[会社名]/'社員'/'@退職者'/[メンバー名] の形式
	entries, err := os.ReadDir(srv.baseDirPath)
	if err != nil {
		return
	}

	// 十分な容量を確保
	targetDirs = make([]string, len(entries)*100)

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

		// 一人親方ディレクトリである場合の処理
		if core.ContainCompanyCategoryName(cat, "一人親方") {
			dirPath := filepath.Join(srv.baseDirPath, entry.Name())
			targetDirs = append(targetDirs, dirPath)
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
				targetDirs = append(targetDirs, dirPath)
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
				targetDirs = append(targetDirs, dirPath)
			}
		}
	}
	return
}

// GetMembers は全ての Member 情報を返します
func (srv *MemberService) GetMembers(
	ctx context.Context,
	req *grpcv1.GetMembersRequest,
) (*grpcv1.GetMembersResponse, error) {
	// キャッシュから全ての Member を取得
	members := make(map[string]*grpcv1.Member, len(srv.cache))
	for id, member := range srv.cache {
		members[id] = member.Member
	}

	response := grpcv1.GetMembersResponse_builder{}.Build()
	response.SetMembers(members)

	return response, nil
}

// GetMember は指定された ID の Member 情報を返します
func (srv *MemberService) GetMember(
	ctx context.Context,
	req *grpcv1.GetMemberRequest,
) (*grpcv1.GetMemberResponse, error) {
	id := req.GetTargetId()

	member, exists := srv.cache[id]
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	response := grpcv1.GetMemberResponse_builder{}.Build()
	response.SetMember(member.Member)

	return response, nil
}

// UpdateMember は Member 情報を更新します
func (srv *MemberService) UpdateMember(
	ctx context.Context,
	req *grpcv1.UpdateMemberRequest,
) (*grpcv1.UpdateMemberResponse, error) {
	targetId := req.GetTargetId()
	sourceMember := req.GetSourceMember()

	if sourceMember == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, nil)
	}

	prevMember, exists := srv.cache[targetId]
	if !exists {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	// prevMemberMessage の作成
	prevMemberMessage := proto.Clone(prevMember.Member).(*grpcv1.Member)

	// source(proto.Message) から Member モデルを作成
	source, err := models.NewMemberFromMessage(sourceMember)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// メンバー情報の更新（マニフェスト保存を含む）
	err = srv.Update(targetId, source)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := grpcv1.UpdateMemberResponse_builder{}.Build()
	response.SetPrevMember(prevMemberMessage)

	return response, nil
}

package models

import (
	"errors"
	"strings"

	"google.golang.org/protobuf/proto"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"
)

// Member は gRPC grpc.v1.Member メッセージの拡張版です。
type Member struct {
	// Member メッセージ本体
	*grpcv1.Member

	// ManifestProvider は Manifestデータの永続化を提供します
	*core.ManifestProvider
}

// NewMember インスタンス作成と初期化を行います
func NewMember(dirPath string) (*Member, error) {
	member := &Member{
		Member: grpcv1.Member_builder{}.Build(),
	}

	err := member.ParseFromDirPath(dirPath)
	if err != nil {
		return nil, err
	}

	// ManifestProvider の初期化
	member.InitializeManifestProvider()

	return member, nil
}

// NewMemberFromMessage は gRPC メッセージから Member インスタンスを生成します
func NewMemberFromMessage(message *grpcv1.Member) (*Member, error) {
	if message == nil {
		return nil, errors.New("message is nil")
	}

	member := &Member{}
	member.Member = proto.Clone(message).(*grpcv1.Member)

	// ManifestProvider の初期化
	member.InitializeManifestProvider()

	return member, nil
}

// InitializeManifestProvider は ManifestProvider を初期化します
func (m *Member) InitializeManifestProvider() {
	mp := &core.ManifestProvider{
		Manifestable:     m,
		ManifestFileName: "@member.yaml",
	}
	m.ManifestProvider = mp
}

// GenerateMemberId は名前から MemberID を生成します
func GenerateMemberIdFromName(dirName string) string {
	return core.GenerateIdFromString(dirName)
}

func (m *Member) ParseFromDirPath(dirPath string) error {
	// dirPath の妥当性チェック
	dirPath, err := core.ResolveAbsPath(dirPath)
	if err != nil {
		return err
	}

	// dirPath をディレクトリディスクリプタで分割
	pathParts := strings.Split(dirPath, "/")

	// メンバーの所属形態を判定
	// .../'1 会社'/'3 個人名' の形式
	// .../'1 会社'/[会社名]/'社員'/[メンバー名] の形式
	// .../'1 会社'/[会社名]/'社員'/'@退職者'/[メンバー名] の形式
	var (
		name                 string
		companyName          string
		companyCategoryIndex int32
	)

	// 一人親方ディレクトリ('3 個人名')かを判定
	// .../'1 会社'/'3 個人名' の形式を想定
	// 少なくとも'1 会社'ディレクトリ直下であることを確認
	if len(pathParts) < 3 {
		return errors.New("ディレクトリパスが短すぎます")
	}

	// パスパーツから'1 会社'ディレクトリの位置を確認
	companyDirectoryIndex := -1
	for i := 1; i < len(pathParts)-1; i++ {
		if pathParts[i] == "1 会社" {
			companyDirectoryIndex = i
			break
		}
	}
	if companyDirectoryIndex == -1 {
		return errors.New("'1 会社'ディレクトリがパスに含まれていません")
	}
	if len(pathParts) < companyDirectoryIndex+2 {
		return errors.New("ディレクトリパスが短すぎます")
	}

	// CompanyCategoryの取得
	companyDirname := pathParts[companyDirectoryIndex+1]
	if len(companyDirname) < 3 || companyDirname[1] != ' ' {
		return errors.New("companyDirnameの形式が規定外です")
	}

	companyCategoryIndex, err = ParseCompanyCategoryByte(companyDirname[0])
	if err != nil {
		return err
	}

	// 会社名の取得
	companyName = companyDirname[2:]

	// 一人親方ディレクトリである場合の処理
	if companyCategoryIndex == 3 {
		// 名前の取得
		m.SetId(GenerateMemberIdFromName(companyDirname))
		m.SetName(companyDirname)
		m.SetCompanyName(CompanyCategoryMap[companyCategoryIndex])
		m.SetIsActive(true)
		m.SetDirPath(dirPath)

		return nil
	}

	// 会社所属メンバーで is_active フラグが true の場合の処理
	// .../'1 会社'/[会社名]/'社員'/[メンバー名] の形式を想定
	if len(pathParts) < companyDirectoryIndex+4 {
		return errors.New("ディレクトリパスが短すぎます")
	}
	if pathParts[companyDirectoryIndex+2] != "社員" {
		return errors.New("社員ディレクトリが存在しません")
	}
	name = pathParts[companyDirectoryIndex+3]

	if name != "@退職者" {
		idPath := strings.Join(pathParts[companyDirectoryIndex+1:companyDirectoryIndex+4], "/")
		m.SetId(GenerateMemberIdFromName(idPath))
		m.SetIsActive(true)

	} else {
		// 退職者ディレクトリの場合の処理
		if len(pathParts) < companyDirectoryIndex+5 {
			return errors.New("ディレクトリパスが短すぎます")
		}
		name = pathParts[companyDirectoryIndex+4]
		idPath := strings.Join(pathParts[companyDirectoryIndex+1:companyDirectoryIndex+5], "/")
		m.SetId(GenerateMemberIdFromName(idPath))
		m.SetIsActive(false)
	}

	// フィールドの設定
	m.SetName(name)
	m.SetCompanyName(companyName)
	m.SetIsActive(false)
	m.SetDirPath(dirPath)

	// id は会社名からメンバー名までのディレクトリ名で生成
	return nil
}

// Update はメンバー情報を更新します
// 必要に応じてメンバーフォルダー名の変更も行います
func (m *Member) Update(source *Member) error {

	// source が nil の場合は マニフェストからデータを読み込む
	if source == nil {
		return m.LoadManifest()
	}

	// Manifest データの更新
	err := m.UpdateManifest(source.ManifestProvider)
	if err != nil {
		return err
	}

	return m.SaveManifest()
}

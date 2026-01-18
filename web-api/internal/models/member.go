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

// Id 生成用テキストを返します
func (m *Member) GetIdSourceText() string {

	return m.GetId()
}

// memberPathInfo はメンバーのパス情報を保持します
type memberPathInfo struct {
	fullPath        string
	companyCategory *grpcv1.CompanyCategory
	companyName     string
	memberName      string
	isActive        bool
	relativePath    []string // "1 会社" 以降の相対パス
}

// ParseFromDirPath はディレクトリパスから Member 情報を解析して設定します
func (m *Member) ParseFromDirPath(dirPath string) error {
	// パスを正規化
	dirPath, err := core.ResolveAbsPath(dirPath)
	if err != nil {
		return err
	}

	// パス情報を抽出
	pathInfo, err := parseMemberPath(dirPath)
	if err != nil {
		return err
	}

	// Member フィールドを設定
	m.SetId(pathInfo.generateId())
	m.SetName(pathInfo.memberName)
	m.SetCompanyName(pathInfo.companyName)
	m.SetCompanyCategoryName(pathInfo.companyCategory.GetName())
	m.SetIsActive(pathInfo.isActive)
	m.SetDirPath(pathInfo.fullPath)

	return nil
}

// parseMemberPath はディレクトリパスを解析して memberPathInfo を返します
func parseMemberPath(dirPath string) (*memberPathInfo, error) {
	parts := strings.Split(dirPath, "/")

	// "1 会社" ディレクトリのインデックスを取得
	companyIdx := findIndex(parts, "1 会社")
	if companyIdx == -1 {
		return nil, errors.New("'1 会社'ディレクトリが見つかりません")
	}

	// "1 会社" 以降の相対パス
	relativePath := parts[companyIdx+1:]
	if len(relativePath) == 0 {
		return nil, errors.New("会社ディレクトリが指定されていません")
	}

	// 会社カテゴリと会社名を取得
	companyCategory, companyName, err := parseCompanyDir(relativePath[0])
	if err != nil {
		return nil, err
	}

	// 一人親方の場合
	if companyCategory.GetName() == "一人親方" {
		return &memberPathInfo{
			fullPath:        dirPath,
			companyCategory: companyCategory,
			companyName:     companyCategory.GetName(),
			memberName:      companyName,
			isActive:        true,
			relativePath:    relativePath[:1],
		}, nil
	}

	// 会社所属メンバーの場合
	return parseCompanyMember(dirPath, relativePath, companyCategory, companyName)
}

// parseCompanyDir は会社ディレクトリ名から会社カテゴリと会社名を抽出します
func parseCompanyDir(dirName string) (*grpcv1.CompanyCategory, string, error) {
	category, err := core.ParseCompanyCategoryFromDirName(dirName)
	if err != nil {
		return nil, "", err
	}

	return category, dirName[2:], nil
}

// parseCompanyMember は会社所属メンバーの情報を解析します
func parseCompanyMember(fullPath string, relativePath []string, category *grpcv1.CompanyCategory, companyName string) (*memberPathInfo, error) {
	// 最低限必要: [会社dir, "社員", メンバー名]
	if len(relativePath) < 3 {
		return nil, errors.New("メンバーパスが不完全です")
	}

	if relativePath[1] != "社員" {
		return nil, errors.New("'社員'ディレクトリが見つかりません")
	}

	// 退職者の場合: [会社dir, "社員", "@退職者", メンバー名]
	if strings.Contains(relativePath[2], "@退職") {
		if len(relativePath) < 4 {
			return nil, errors.New("退職者のメンバー名が指定されていません")
		}
		return &memberPathInfo{
			fullPath:        fullPath,
			companyCategory: category,
			companyName:     companyName,
			memberName:      relativePath[3],
			isActive:        false,
			relativePath:    relativePath[:4],
		}, nil
	}

	// 在籍メンバーの場合: [会社dir, "社員", メンバー名]
	return &memberPathInfo{
		fullPath:        fullPath,
		companyCategory: category,
		companyName:     companyName,
		memberName:      relativePath[2],
		isActive:        true,
		relativePath:    relativePath[:3],
	}, nil
}

// generateId は ID 生成用の文字列から ID を生成します
func (p *memberPathInfo) generateId() string {
	return GenerateMemberIdFromName(strings.Join(p.relativePath, "/"))
}

// findIndex はスライス内で指定された値のインデックスを返します
func findIndex(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
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

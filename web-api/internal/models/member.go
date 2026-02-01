package models

import (
	"errors"
	"strings"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

// Member は core.PersistModel[*grpcv1.Member] の拡張版です。
type Member struct{}

// GenerateId は dirPath から会社IDを生成します
func (m *Member) GenerateId(message protoreflect.Message) string {
	pathInfo, err := messageToMemberPathInfo(message)
	if err != nil {
		return ""
	}
	return core.BytesToId([]byte(pathInfo.relativePath))
}

// memberPathInfo はメンバーのパス情報を保持します
type memberPathInfo struct {
	fullPath        string
	companyCategory *grpcv1.CompanyCategory
	companyName     string
	memberName      string
	isActive        bool
	relativePath    string // "1 会社" 以降の相対パス
}

// ParseFromDirPath はディレクトリパスから Member 情報を解析して設定します
func (m *Member) GenerateMessage(request protoreflect.Message) (protoreflect.ProtoMessage, error) {
	// request が nil の場合はデフォルト初期化を行う
	if request == nil {
		return grpcv1.Member_builder{}.Build(), nil
	}

	// パス情報を抽出
	pathInfo, err := messageToMemberPathInfo(request)
	if err != nil {
		return nil, err
	}

	// Member メッセージの生成
	mes := grpcv1.Member_builder{}.Build()

	// Member フィールドを設定
	mes.SetName(pathInfo.memberName)
	mes.SetCompanyName(pathInfo.companyName)
	mes.SetCompanyCategoryName(pathInfo.companyCategory.GetName())
	mes.SetIsActive(pathInfo.isActive)
	mes.SetDirPath(pathInfo.fullPath)

	newId := m.GenerateId(mes.ProtoReflect())
	mes.SetId(newId)

	return mes, nil
}

// messageToMemberPathInfo はディレクトリパスを解析して memberPathInfo を返します
func messageToMemberPathInfo(request protoreflect.Message) (*memberPathInfo, error) {
	// request の型アサーション
	req, ok := request.Interface().(*grpcv1.Member)
	if !ok {
		return nil, errors.New("message の型アサーションに失敗しました")
	}

	// dirPath を取得する
	dirPath := req.GetDirPath()

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
			relativePath:    strings.Join(relativePath[:1], "/"),
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
			relativePath:    strings.Join(relativePath[:4], "/"),
		}, nil
	}

	// 在籍メンバーの場合: [会社dir, "社員", メンバー名]
	return &memberPathInfo{
		fullPath:        fullPath,
		companyCategory: category,
		companyName:     companyName,
		memberName:      relativePath[2],
		isActive:        true,
		relativePath:    strings.Join(relativePath[:3], "/"),
	}, nil
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
func (m *Member) UpdateMessage(target protoreflect.Message, source protoreflect.Message) error {
	// target と source の型アサーション
	_, ok1 := target.Interface().(*grpcv1.Member)
	_, ok2 := source.Interface().(*grpcv1.Member)
	if !ok1 || !ok2 {
		return errors.New("message の型アサーションに失敗しました")
	}

	// Manifest データの更新
	// TODO: メンバー情報の更新ロジックを実装

	return nil
}

// NewPersistModelMember は指定されたメンバーフォルダーから PersistModel[*Member] を作成します
func NewPersistModelMember(dirPath string) (*core.PersistModel[*Member], error) {
	// PersistModel を作成
	pm, err := core.NewPersistModel(&Member{}, "@member.yaml")
	if err != nil {
		return nil, err
	}

	// 初期化
	request := grpcv1.Member_builder{}.Build()
	request.SetDirPath(dirPath)
	err = pm.Initialize(request.ProtoReflect())
	if err != nil {
		return nil, err
	}

	//
	return pm, nil
}

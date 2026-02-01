package models

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"
)

// Company は core.PersistModel[*grpcv1.Company] の拡張版です。
type Company struct{}

// GenerateMessage はモデルの protobuf メッセージを取得します。
func (m *Company) GenerateMessage(request protoreflect.Message) (protoreflect.ProtoMessage, error) {
	// mes が 空文字 の場合はデフォルト初期化を行う
	if request == nil {
		return grpcv1.Company_builder{}.Build(), nil
	}

	// message の型アサーション
	req, ok := request.Interface().(*grpcv1.Company)
	if !ok {
		return nil, errors.New("message の型アサーションに失敗しました")
	}

	// dirPath を取得する
	dirPath := req.GetDirPath()

	// ParseFromDirPath は"[0-9] [会社名]"形式のファイル名となっているパスを解析します
	// 会社名内のハイフン（含まれる場合）以前の文字列を会社名、ハイフン以降の文字列を関連名として扱います
	// 戻り値Companyは: Id, Target, Cateory, Name, Tags のみ設定されます

	// ディレクトリ名の取得
	dirName := core.GetBaseName(dirPath)

	// ディレクトリ名解析
	// 3文字以上のdirNameかチェック
	if len(dirName) < 3 || dirName[1] != ' ' {
		return nil, errors.New("dirNameの形式が規定外です")
	}

	// 会社名の取得
	sn := dirName[2:]
	if sn == "" {
		return nil, errors.New("会社名部分が取得できません")
	}

	// CategoryIndexの取得
	cat, err := core.ParseCompanyCategoryFromDirName(dirName)
	if err != nil {
		return nil, err
	}

	// Companyメッセージの生成
	mes := grpcv1.Company_builder{}.Build()

	// 各フィールドの設定
	mes.SetId(m.GenerateId(request))
	mes.SetDirPath(dirPath)
	mes.SetCategoryIndex(cat.GetIndex())
	mes.SetName(sn)

	return mes, nil
}

// GenerateId は dirPath から会社IDを生成します
func (m *Company) GenerateId(message protoreflect.Message) string {
	mes, ok := message.Interface().(*grpcv1.Company)
	if !ok {
		return ""
	}
	dirPath := mes.GetDirPath()
	basename := core.GetBaseName(dirPath)
	return core.BytesToId([]byte(basename))
}

// Update は会社情報を更新します
// 必要に応じて会社フォルダー名の変更も行います
func (m *Company) UpdateMessage(target protoreflect.Message, source protoreflect.Message) error {

	// source メッセージの型アサーション
	srcMes, ok := source.Interface().(*grpcv1.Company)
	if !ok {
		return errors.New("source メッセージの型アサーションに失敗しました")
	}

	// 新しいパラメータを元に会社フォルダーパスを生成
	baseName := m.generateBaseName(
		srcMes.GetCategoryIndex(),
		srcMes.GetName(),
	)
	if baseName == "" {
		return errors.New("新しい会社フォルダー名の取得に失敗しました")
	}

	// target メッセージの型アサーション
	mes, ok := target.Interface().(*grpcv1.Company)
	if !ok {
		return errors.New("target メッセージの型アサーションに失敗しました")
	}

	// 会社フォルダーパスの生成
	parentPath := filepath.Dir(mes.GetDirPath())
	dirPath := filepath.Join(parentPath, baseName)

	// ファイル名変更の必要がある場合は会社フォルダー名を更新
	if dirPath != mes.GetDirPath() {

		// フォルダー名変更
		err := os.Rename(mes.GetDirPath(), dirPath)
		if err != nil {
			return err
		}

		// フィールドの更新
		mes.SetDirPath(dirPath)
		mes.SetCategoryIndex(srcMes.GetCategoryIndex())
		mes.SetName(srcMes.GetName())

		// Id フィールドの更新
		newId := m.GenerateId(mes.ProtoReflect())
		mes.SetId(newId)

	}

	return nil
}

// generateBaseName はパラメータをもとに会社フォルダー名変更します
//
//	引数：
//	  ci: カテゴリーインデックス
//	  sn: 省略会社名
//
//	戻り値: 生成された会社フォルダーパス
func (m *Company) generateBaseName(ci int32, sn string) string {
	dirName := strconv.Itoa(int(ci)) + " " + sn
	return dirName
}

// NewPersistModelCompany は指定された会社フォルダーから PersistModel[*Company] を作成します
func NewPersistModelCompany(dirPath string) (*core.PersistModel[*Company], error) {
	// PersistModel を作成
	pm, err := core.NewPersistModel(&Company{}, "company.yaml")
	if err != nil {
		return nil, err
	}

	// 初期化
	request := grpcv1.Company_builder{}.Build()
	request.SetDirPath(dirPath)
	err = pm.Initialize(request.ProtoReflect())
	if err != nil {
		return nil, err
	}

	//
	return pm, nil
}

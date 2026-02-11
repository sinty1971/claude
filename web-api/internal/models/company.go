package models

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"
)

// Company は core.PersistModel の型安全な実装です。
// 新規実装ではこちらを使用してください。
type Company struct{}

// InitializeFromMessage は message メッセージを元に、ファイルシステム情報を反映した protobuf メッセージを構築します。
func (m *Company) InitializeFromMessage(message *grpcv1.Company) (*grpcv1.Company, error) {
	// message が nil の場合はデフォルト初期化を行う
	if message == nil {
		return grpcv1.Company_builder{}.Build(), nil
	}

	// dirPath を取得する
	dirPath := message.GetDirPath()

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
	mes.SetId(m.generateId(message))
	mes.SetDirPath(dirPath)
	mes.SetCategoryIndex(cat.GetIndex())
	mes.SetName(sn)

	return mes, nil
}

// UpdateMessage は source の値に基づいて target メッセージを更新します。
func (m *Company) UpdateMessage(target *grpcv1.Company, source *grpcv1.Company) error {
	// 新しいパラメータを元に会社フォルダーパスを生成
	baseName := m.generateBaseName(
		source.GetCategoryIndex(),
		source.GetName(),
	)
	if baseName == "" {
		return errors.New("新しい会社フォルダー名の取得に失敗しました")
	}

	// 会社フォルダーパスの生成
	parentPath := filepath.Dir(target.GetDirPath())
	dirPath := filepath.Join(parentPath, baseName)

	// ファイル名変更の必要がある場合は会社フォルダー名を更新
	if dirPath != target.GetDirPath() {
		// フォルダー名変更
		err := os.Rename(target.GetDirPath(), dirPath)
		if err != nil {
			return err
		}

		// フィールドの更新
		target.SetDirPath(dirPath)
		target.SetCategoryIndex(source.GetCategoryIndex())
		target.SetName(source.GetName())

		// Id フィールドの更新
		newId := m.generateId(target)
		target.SetId(newId)
	}

	return nil
}

// generateId は dirPath から会社IDを生成します
func (m *Company) generateId(message *grpcv1.Company) string {
	dirPath := message.GetDirPath()
	basename := core.GetBaseName(dirPath)
	return core.BytesToId([]byte(basename))
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

// NewPersistModelCompany は指定された会社フォルダーから PersistModel[*Company] を作成します。
// 新規実装ではこちらを使用してください。
func NewPersistModelCompany(dirPath string) (*core.PersistModel[*grpcv1.Company, *Company], error) {
	// PersistModel を作成
	pm, err := core.NewPersistModel(&Company{}, "@company.yaml")
	if err != nil {
		return nil, err
	}

	// 初期化
	request := grpcv1.Company_builder{}.Build()
	request.SetDirPath(dirPath)
	err = pm.Initialize(request)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

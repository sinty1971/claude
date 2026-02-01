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

// Company は gRPC grpc.v1.Company メッセージの拡張版です。
type Company struct{}

// GetMessage はモデルの protobuf メッセージを取得します。
func (m *Company) GenerateMessage(initArg string) (protoreflect.ProtoMessage, error) {
	// initArg が 空文字 の場合はデフォルト初期化を行う
	if initArg == "" {
		return grpcv1.Company_builder{}.Build(), nil
	}

	// initArg を dirPath として progobuf メッセージを取得する
	return m.MessageFromDirPath(initArg)
}

// GenerateId は dirPath から会社IDを生成します
func (m *Company) GenerateId(dirName string) string {
	basename := core.GetBaseName(dirName)
	return core.GenerateIdFromString(basename)
}

// ParseFromDirPath は"[0-9] [会社名]"形式のファイル名となっているパスを解析します
// 会社名内のハイフン（含まれる場合）以前の文字列を会社名、ハイフン以降の文字列を関連名として扱います
// 戻り値Companyは: Id, Target, Cateory, Name, Tags のみ設定されます
func (m *Company) MessageFromDirPath(dirPath string) (protoreflect.ProtoMessage, error) {

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

	mes := grpcv1.Company_builder{}.Build()

	// 各フィールドの設定
	mes.SetId(m.GenerateId(dirName))
	mes.SetDirPath(dirPath)
	mes.SetCategoryIndex(cat.GetIndex())
	mes.SetName(sn)

	return mes, nil
}

// Update は会社情報を更新します
// 必要に応じて会社フォルダー名の変更も行います
func (m *Company) Update(target *core.PersistModel[*Company], source *core.PersistModel[*Company]) error {

	// source が nil の場合は m.dirPath から再解析を行う
	if source == nil {
		// 会社フォルダーパスから再解析
		targetDirPath, err := target.GetDirPath()
		if err != nil {
			return err
		}
		err = target.BuildWithInitArg(targetDirPath)
		if err != nil {
			return err
		}
		return target.Load()
	}

	// 新しい会社フォルダー名の取得
	srcMes := source.Message.Interface().(*grpcv1.Company)
	newDirName := m.generateDirName(
		srcMes.GetCategoryIndex(),
		srcMes.GetName(),
	)
	if newDirName == "" {
		return errors.New("新しい会社フォルダー名の取得に失敗しました")
	}

	// 新しいパラメータを元に会社フォルダーパスを生成
	trgMes := target.Message.Interface().(*grpcv1.Company)
	newDirPath := filepath.Join(filepath.Dir(trgMes.GetDirPath()), newDirName)

	// ファイル名変更の必要がある場合は会社フォルダー名を更新
	if trgMes.GetDirPath() != newDirPath {

		// フォルダー名変更
		err := os.Rename(trgMes.GetDirPath(), newDirPath)
		if err != nil {
			return err
		}

		// newId の取得
		newId := m.GenerateId(newDirName)

		// フィールドの更新
		// マニフェスト以外の情報を更新
		trgMes.SetId(newId)
		trgMes.SetDirPath(newDirPath)
		trgMes.SetCategoryIndex(srcMes.GetCategoryIndex())
		trgMes.SetName(srcMes.GetName())
	}

	// Manifest データの更新
	err := target.UpdateManifest(source)
	if err != nil {
		return err
	}

	return m.Save()
}

// generateDirPath はパラメータをもとに会社フォルダー名変更します
//
//	引数：
//	  ci: カテゴリーインデックス
//	  sn: 省略会社名
//
//	戻り値: 生成された会社フォルダーパス
func (m *Company) generateDirName(ci int32, sn string) string {
	dirName := strconv.Itoa(int(ci)) + " " + sn
	return dirName
}

// Save はマニフェストを保存します（Manifestable インターフェース実装）
func (m *Company) Save() error {
	if m.ManifestProvider == nil {
		return errors.New("ManifestProvider is nil")
	}
	return m.ManifestProvider.Save()
}

// Load はマニフェストを読み込みます（Manifestable インターフェース実装）
func (m *Company) Load() error {
	if m.ManifestProvider == nil {
		return errors.New("ManifestProvider is nil")
	}
	return m.ManifestProvider.Load()
}

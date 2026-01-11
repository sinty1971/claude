package models

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"google.golang.org/protobuf/proto"

	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"
)

// Company は gRPC grpc.v1.Company メッセージの拡張版です。
type Company struct {
	// Company メッセージ本体
	*grpcv1.Company

	// ManifestProvider は Manifestデータの永続化設定を管理します
	*core.ManifestProvider
}

// NewCompany インスタンス作成と初期化を行います
func NewCompany() *Company {

	// インスタンス作成と初期化
	company := &Company{}
	company.Company = grpcv1.Company_builder{}.Build()
	company.ManifestProvider = core.NewManifestProvider(company)

	return company
}

// GetManifestFolder は Manifest ファイルを保存先フルパスを取得します
// Manifestable インターフェースの実装
func (m *Company) GetManifestFolder() string {
	return m.GetDirPath()
}

// GetProtoMessage は Company の protobuf メッセージを取得します
// Manifestable インターフェースの実装
func (m *Company) GetProtoMessage() proto.Message {
	return m.Company
}

// ParseFromPath は"[0-9] [会社名]"形式のファイル名となっているパスを解析します
// 会社名内のハイフン（含まれる場合）以前の文字列を会社名、ハイフン以降の文字列を関連名として扱います
// 戻り値Companyは: Id, Target, Cateory, ShortName, Tags のみ設定されます
func (m *Company) ParseFromPath(paths ...string) error {

	// パスを結合
	dirPath := filepath.Join(paths...)

	// ディレクトリ名の取得
	dirName := filepath.Base(dirPath)

	// ディレクトリ名解析
	ci, sn, err := m.parseDirName(dirName)
	if err != nil {
		return err
	}

	id := core.GenerateIdFromString(dirName)

	// 各フィールドの設定
	m.SetId(id)
	m.SetDirPath(dirPath)
	m.SetCategoryIndex(int32(ci))
	m.SetShortName(sn)

	return nil
}

// parseDirName は会社ディレクトリであろうディレクトリ名の解析
//
//	dirName: Directory Name
//	ci: CategoryIndex
//	sn: 会社名の略称
//	err: エラー情報
func (m *Company) parseDirName(dirName string) (ci int32, sn string, err error) {
	// 3文字以上のdirNameかチェック
	if len(dirName) < 3 || dirName[1] != ' ' {
		return -1, "", errors.New("dirNameの形式が規定外です")
	}

	// 会社名部分の取得
	sn = dirName[2:]
	if sn == "" {
		return -1, "", errors.New("会社名部分が取得できません")
	}

	// CategoryIndexの取得
	if idx, err := strconv.Atoi(string(dirName[0])); err != nil {
		return -1, "", err
	} else {
		ci = int32(idx)
	}

	// CategoryIndexの妥当性チェック
	if err := ErrorCompanyCategoryIndex(ci); err != nil {
		return -1, "", err
	}

	return ci, sn, nil
}

// Update は会社情報を更新します
// 必要に応じて管理フォルダー名の変更も行います
func (m *Company) Update(source *Company) error {

	// 引数チェック
	if source == nil {
		return errors.New("更新情報 source の値が nil です")
	}

	// 新しいパラメータを元に管理フォルダーパスを生成
	newDirPath := m.GenerateDirPath(
		filepath.Dir(m.GetDirPath()),
		source.GetCategoryIndex(),
		source.GetShortName(),
	)

	// ファイル名変更の必要がある場合は管理フォルダー名を更新
	if m.GetDirPath() != newDirPath {

		// フォルダー名変更
		if err := os.Rename(m.GetDirPath(), newDirPath); err != nil {
			return err
		}
	}

	// Persist情報の更新
	return m.ManifestProvider.Update(source.ManifestProvider)
}

// GenerateCompanyPath はパラメータをもとに管理フォルダー名変更します
//
// 引数:
//
//	dir: 基本パス(原則として O:/.../1 会社 などの親フォルダー)
//	ci: カテゴリーインデックス
//	sn: 省略会社名
//
// 戻り値:
//
//	生成された管理フォルダーパス
func (m *Company) GenerateDirPath(dir string, ci int32, sn string) string {
	dirName := strconv.Itoa(int(ci)) + " " + sn
	return filepath.Join(dir, dirName)
}

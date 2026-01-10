package models

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
func (m *Company) GetManifestFolder() string {
	return m.GetDirPath()
}

func (m *Company) GetProtoMessage() proto.Message {
	return m.Company
}

// GetId は会社の一意なIDを生成します
// IDは会社フォルダー名から生成されます
func (m *Company) GenerateId() string {

	text := filepath.Base(m.GetDirPath())
	return core.GenerateIdFromString(text)
}

// ParseFrom は"[0-9] [会社名]"形式のファイル名となっているパスを解析します
// 会社名内のハイフン（含まれる場合）以前の文字列を会社名、ハイフン以降の文字列を関連名として扱います
// 戻り値Companyは: Id, Target, Cateory, ShortName, Tags のみ設定されます
func (m *Company) ParseFrom(paths ...string) error {

	// パスを結合
	dirPath := filepath.Join(paths...)

	// 引数 target からフォルダー名取得とチェック
	// "[0-9] [会社名]"の解析
	dirName := filepath.Base(dirPath)
	if len(dirName) < 3 {
		return errors.New("targetのファイル名形式が無効です（長さが短い）")
	} else if dirName[1] != ' ' {
		// 2番目の文字がスペースかチェック
		return errors.New("targetaのファイル名形式が無効です")
	}

	// カテゴリー情報の取得
	var catIndex int
	var err error

	if catIndex, err = strconv.Atoi(string(dirName[0])); err != nil {
		return err
	}
	if err := ErrorCompanyCategoryIndex(catIndex); err != nil {
		return err
	}

	// 会社フォルダー名の解析
	nameParts := strings.Split(dirName[2:], " ")
	if len(nameParts) == 0 || nameParts[0] == "" {
		return errors.New("会社名が取得できません")
	}

	// 会社名の解析（ハイフンで分割）
	shortName := nameParts[0]
	if idx := strings.Index(nameParts[0], "-"); idx > -1 {
		// ハイフン以前の文字列を会社名とする
		shortName = nameParts[0][:idx]
		// ハイフン以降の文字列を関連文字列とする
	}

	// Target,Category,ShortNameの設定
	m.SetDirPath(dirPath)
	m.SetCategoryIndex(int32(catIndex))
	m.SetShortName(shortName)

	// IDの設定、targetの設定が終了した後に実行
	m.SetId(m.GenerateId())
	return nil
}

// Update は会社情報を更新します
// 必要に応じて管理フォルダー名の変更も行います
func (m *Company) Update(source *Company) error {

	// 引数チェック
	if source == nil {
		return errors.New("source Company is nil")
	}

	// 新しいパラメータを元に管理フォルダーパスを生成
	newDirPath := GenerateCompanyDirPath(
		filepath.Dir(m.GetDirPath()),
		source.GetCategoryIndex(),
		source.GetShortName(),
	)

	// ファイル名変更の必要がある場合は管理フォルダー名を更新
	if newDirPath != m.GetDirPath() {

		// フォルダー名変更
		if err := os.Rename(m.GetDirPath(), newDirPath); err != nil {
			return err
		}
	}

	// Persist情報の更新
	return m.ManifestProvider.Update(source.ManifestProvider)
}

// GenerateCompanyDirPath はパラメータをもとに管理フォルダー名変更します
// parentPath: 基本パス(原則として　O:/.../1 会社 などの親ディレクトリパス)
// idx: カテゴリーインデックス
// name: 省略会社名
func GenerateCompanyDirPath(parentPath string, idx int32, name string) string {
	DirName := strconv.Itoa(int(idx)) + " " + name
	return filepath.Join(parentPath, DirName)
}

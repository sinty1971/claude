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

	// ManifestProvider は Manifestデータの永続化を提供します
	Manifest *core.ManifestProvider
}

// NewCompany インスタンス作成と初期化を行います
// Manifest は初期化をしますが Manifest ファイルの読み込みは行いません
func NewCompany(dirPath string) (*Company, error) {

	// インスタンス作成と初期化
	company := &Company{}
	company.Company = grpcv1.Company_builder{}.Build()

	if dirPath == "" {
		return nil, errors.New("dirPath is empty")
	}

	// dirPath から情報を解析して設定
	err := company.ParseFromDirPath(dirPath)
	if err != nil {
		return nil, err
	}

	// ManifestProvider の初期化
	company.Manifest = core.NewManifestProvider(company)

	return company, nil
}

func NewCompanyFromMessage(msg *grpcv1.Company) (*Company, error) {
	if msg == nil {
		return nil, errors.New("msg is nil")
	}

	// クローンを作成して保持
	cloneMsg := proto.Clone(msg).(*grpcv1.Company)
	company := &Company{Company: cloneMsg}

	// ManifestProvider の初期化
	company.Manifest = core.NewManifestProvider(company)

	return company, nil
}

// GenerateCompanyId は dirPath から会社IDを生成します
func GenerateCompanyIdFromDirName(dirName string) string {
	return core.GenerateIdFromString(dirName)
}

// GetManifestFolder は Manifest ファイルを保存先フルパスを取得します
// Manifestable インターフェースの実装
func (m *Company) GetManifestDirectory() string {
	return m.GetDirPath()
}

// GetManifestMessage は Company の protobuf メッセージを取得します
// Manifestable インターフェースの実装
func (m *Company) GetManifestMessage() proto.Message {
	return m.Company
}

// ParseFromDirPath は"[0-9] [会社名]"形式のファイル名となっているパスを解析します
// 会社名内のハイフン（含まれる場合）以前の文字列を会社名、ハイフン以降の文字列を関連名として扱います
// 戻り値Companyは: Id, Target, Cateory, ShortName, Tags のみ設定されます
func (m *Company) ParseFromDirPath(dirPath string) error {

	// ディレクトリ名の取得
	dirName := core.GetBaseName(dirPath)

	// ディレクトリ名解析
	// 3文字以上のdirNameかチェック
	if len(dirName) < 3 || dirName[1] != ' ' {
		return errors.New("dirNameの形式が規定外です")
	}

	// 会社名の取得
	sn := dirName[2:]
	if sn == "" {
		return errors.New("会社名部分が取得できません")
	}

	// CategoryIndexの取得
	num, err := strconv.Atoi(string(dirName[0]))
	if err != nil {
		return err
	}
	ci := int32(num)

	// CategoryIndexの妥当性チェック
	err = ErrorCompanyCategoryIndex(ci)
	if err != nil {
		return err
	}

	// 各フィールドの設定
	m.SetId(GenerateCompanyIdFromDirName(dirName))
	m.SetDirPath(dirPath)
	m.SetCategoryIndex(int32(ci))
	m.SetShortName(sn)

	return nil
}

// Update は会社情報を更新します
// 必要に応じて会社フォルダー名の変更も行います
func (m *Company) Update(source *Company) error {

	// source が nil の場合は m.dirPath から再解析を行う
	if source == nil {
		err := m.ParseFromDirPath(m.GetDirPath())
		if err != nil {
			return err
		}
		return m.Manifest.Load()
	}

	// 新しいパラメータを元に会社フォルダーパスを生成
	newDirPath, err := m.GenerateDirPath(
		filepath.Dir(m.GetDirPath()),
		source.GetCategoryIndex(),
		source.GetShortName(),
	)
	if err != nil {
		return err
	}

	// 新しい会社フォルダー名の取得
	newDirName := core.GetBaseName(newDirPath)
	if newDirName == "" {
		return errors.New("新しい会社フォルダー名の取得に失敗しました")
	}

	// ファイル名変更の必要がある場合は会社フォルダー名を更新
	if m.GetDirPath() != newDirPath {

		// フォルダー名変更
		err := os.Rename(m.GetDirPath(), newDirPath)
		if err != nil {
			return err
		}

		// newId の取得
		newId := GenerateCompanyIdFromDirName(newDirName)

		// フィールドの更新
		// マニフェスト以外の情報を更新
		m.SetId(newId)
		m.SetDirPath(newDirPath)
		m.SetCategoryIndex(source.GetCategoryIndex())
		m.SetShortName(source.GetShortName())
	}

	// Manifest データの更新
	err = m.Manifest.Update(source.Manifest)
	if err != nil {
		return err
	}
	return m.Manifest.Save()
}

// GenerateDirPath はパラメータをもとに会社フォルダー名変更します
//
//	引数： dir: 基本パス(原則として O:/.../1 会社 などの親フォルダー)
//	 ci: カテゴリーインデックス
//	 sn: 省略会社名
//
//	 戻り値:
//
//		生成された会社フォルダーパス
func (m *Company) GenerateDirPath(dir string, ci int32, sn string) (string, error) {
	dirName := strconv.Itoa(int(ci)) + " " + sn
	return filepath.Join(dir, dirName), nil
}

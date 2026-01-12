package models

import (
	"errors"
	"os"
	"strings"
	"time"
	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"

	"google.golang.org/protobuf/proto"
)

type Koji struct {
	// Koji メッセージ本体
	*grpcv1.Koji

	// Persist共通モデルフィールド
	Manifest *core.ManifestProvider
}

// NewKoji FolderNameからKojiを作成します（高速化版）
func NewKoji(dirPath string) (*Koji, error) {

	koji := &Koji{}
	koji.Koji = grpcv1.Koji_builder{}.Build()
	err := koji.ParseFromDirPath(dirPath)
	if err != nil {
		return nil, err
	}

	// ManifestProvider の初期化
	koji.Manifest = core.NewManifestProvider(koji)

	return koji, nil
}

func NewKojiFromMessage(msg *grpcv1.Koji) (*Koji, error) {
	if msg == nil {
		return nil, errors.New("msg is nil")
	}
	cloneMsg := proto.Clone(msg).(*grpcv1.Koji)

	koji := &Koji{Koji: cloneMsg}

	// ManifestProvider の初期化
	koji.Manifest = core.NewManifestProvider(koji)
	return koji, nil
}

// GetManifestDirectory は Manifest ファイルを保存先フルパスを取得します
// Manifestable インターフェースの実装
func (m *Koji) GetManifestDirectory() string {
	return m.GetDirPath()
}

// GetManifestMessage は Koji の protobuf メッセージを取得します
// Manifestable インターフェースの実装
func (m *Koji) GetManifestMessage() proto.Message {
	if m == nil {
		return nil
	}
	return m.Koji
}

// ParseFromDirPath は dirPath から工事開始日・会社名・現場名を取得
func (m *Koji) ParseFromDirPath(dirPath string) error {

	// フォルダー名を取得
	dirName := core.GetBaseName(dirPath)

	// ファイル名から工事開始日の取得と日付除外文字列の取得
	start := new(Timestamp)
	dirNameRemovedDate, err := ParseTimestamp(dirName, start)
	if err != nil || dirNameRemovedDate == "" {
		return errors.New("工事フォルダー名から工事開始日が取得できません error: " + err.Error())
	}

	// 会社名と現場名の取得
	var companyName string
	var locationName string

	// 最初のスペースで分割（最適化）
	if idx := strings.Index(dirNameRemovedDate, " "); idx > 0 {
		companyName = dirNameRemovedDate[:idx]
		if idx+1 < len(dirNameRemovedDate) {
			locationName = dirNameRemovedDate[idx+1:]
		}
	} else {
		return errors.New("工事フォルダー名から会社名及び現場名が得できません")
	}

	// IDの生成
	id := core.GenerateIdFromString(dirName)

	// 各フィールドの設定
	m.SetId(id)
	m.SetDirPath(dirPath)
	m.SetStart(start.Timestamp)
	m.SetCompanyName(companyName)
	m.SetLocationName(locationName)

	return nil
}

// GenerateKojiStatus はプロジェクトステータスを判定する
func GenerateKojiStatus(start *Timestamp, end *Timestamp) string {
	if start == nil || end == nil {
		return "不明"
	}

	now := time.Now()

	if !start.IsValid() || !end.IsValid() {
		return "不明"
	} else if now.Before(start.AsTime()) {
		return "予定"
	} else if now.After(end.AsTime()) {
		return "完了"
	} else {
		return "進行中"
	}
}

func (m *Koji) Update(source *Koji) error {
	// 引数チェック
	if source == nil {
		return errors.New("更新情報 source の値が nil です")
	}

	// 新しいパラメータを元に管理フォルダーパスを生成
	sourceStart := Timestamp{Timestamp: source.GetStart()}
	newDirPath, err := m.GenerateDirPath(
		sourceStart,
		source.GetCompanyName(),
		source.GetLocationName())
	if err != nil {
		return err
	}

	// フォルダー名変更が必要な場合
	if m.GetDirPath() != newDirPath {
		err := os.Rename(m.GetDirPath(), newDirPath)
		if err != nil {
			return err
		}

		// マニフェスト以外の情報を更新
		m.SetDirPath(newDirPath)
		m.SetStart(source.GetStart())
		m.SetCompanyName(source.GetCompanyName())
		m.SetLocationName(source.GetLocationName())
		m.GenerateId()

	}

	// Persist情報の更新
	return m.Manifest.Update(source.Manifest)
}

func (m *Koji) GenerateId() string {
	dirName := core.GetBaseName(m.GetDirPath())
	id := core.GenerateIdFromString(dirName)
	m.SetId(id)
	return id
}

func (m *Koji) GenerateDirPath(st Timestamp, cn string, loc string) (string, error) {

	startText, err := st.FormatTime("2006-0102")
	if err != nil {
		return "", err
	}

	// 事前に容量を計算してstrings.Builderを初期化（再アロケーション回避）
	// 日付(9文字) + スペース(1文字) + 会社名 + スペース(1文字) + 現場名 の概算
	var dirPathBuilder strings.Builder
	dirPathBuilder.Grow(len(startText) + 1 + len(cn) + 1 + len(loc))

	// 日付部分を手動構築（YYYY-MMDD形式）
	dirPathBuilder.WriteString(startText)

	// 会社名と現場名を追加
	dirPathBuilder.WriteByte(' ')
	dirPathBuilder.WriteString(cn)
	dirPathBuilder.WriteByte(' ')
	dirPathBuilder.WriteString(loc)
	return dirPathBuilder.String(), nil
}

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
	*core.ManifestProvider
}

// NewKoji FolderNameからKojiを作成します
func NewKoji(dirPath string) (*Koji, error) {
	// インスタンス作成
	koji := &Koji{}

	// Koji メッセージ本体の初期化
	koji.Koji = grpcv1.Koji_builder{}.Build()

	// dirPath から情報を解析して設定
	err := koji.ParseFromDirPath(dirPath)
	if err != nil {
		return nil, err
	}

	// ManifestProvider の初期化
	koji.InitializeManifestProvider()

	return koji, nil
}

func NewKojiFromMessage(message *grpcv1.Koji) (*Koji, error) {
	// message が nil の場合はエラーを返す
	if message == nil {
		return nil, errors.New("msg is nil")
	}

	// インスタンス作成
	koji := &Koji{}
	koji.Koji = proto.Clone(message).(*grpcv1.Koji)

	// ManifestProvider の初期化
	koji.InitializeManifestProvider()

	return koji, nil
}

func (m *Koji) InitializeManifestProvider() {
	mp := &core.ManifestProvider{Manifestable: m}
	m.ManifestProvider = mp
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

	// 各フィールドの設定
	m.SetId(GenerateKojiIdFromDirName(dirName))
	m.SetDirPath(dirPath)
	m.SetStart(start.Timestamp)
	m.SetCompanyName(companyName)
	m.SetLocationName(locationName)

	return nil
}

func GenerateKojiIdFromDirName(dirName string) string {
	return core.GenerateIdFromString(dirName)
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
	// source が nil の場合は自身の dirPath から再解析を行う
	if source == nil {
		err := m.ParseFromDirPath(m.GetDirPath())
		if err != nil {
			return err
		}
		return m.LoadManifest()
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

	newDirName := core.GetBaseName(newDirPath)
	if newDirName == "" {
		return errors.New("新しい工事フォルダー名の取得に失敗しました")
	}

	// ディレクトリ名変更が必要な場合
	if m.GetDirPath() != newDirPath {
		err := os.Rename(m.GetDirPath(), newDirPath)
		if err != nil {
			return err
		}

		newId := GenerateKojiIdFromDirName(newDirName)

		// マニフェスト以外の情報を更新
		m.SetId(newId)
		m.SetDirPath(newDirPath)
		m.SetStart(source.GetStart())
		m.SetCompanyName(source.GetCompanyName())
		m.SetLocationName(source.GetLocationName())

	}

	// Manifestデータの更新
	err = m.UpdateManifest(source.ManifestProvider)
	if err != nil {
		return err
	}

	m.EnsureMfEndFromStart()

	return m.SaveManifest()
}

// GenerateDirPath はパラメータをもとに工事フォルダー名変更します
//
//	引数： st: 工事開始日
//	      cn: 会社名
//	      loc: 現場名
//	戻り値:
//
//		生成された工事フォルダーパス
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

// LoadManifest は Manifest ファイルから永続化データを読み込みます
func (m *Koji) LoadManifest() error {
	err := m.ManifestProvider.LoadManifest()
	if err != nil {
		return err
	}

	updated := m.EnsureMfEndFromStart()
	if updated {
		return m.SaveManifest()
	}
	return nil
}

// EnsureMfEndFromStart は mf_end が空の場合に start の値をコピーします
func (m *Koji) EnsureMfEndFromStart() bool {

	if m == nil || m.Koji == nil {
		return false
	}

	// mf_end が有効な場合は終了
	if m.GetMfEnd() != nil && m.GetMfEnd().IsValid() {
		return false
	}

	// start の無効な場合は終了
	if m.GetStart() == nil || !m.GetStart().IsValid() {
		return false
	}

	// start の値を mf_end にコピー
	m.SetMfEnd(m.GetStart())
	return true
}

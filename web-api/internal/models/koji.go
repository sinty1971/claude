package models

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"
)

// Koji は core.PersistModel の型安全な実装です。
// 新規実装ではこちらを使用してください。
type Koji struct{}

// InitializeFromMessage は message メッセージを元に、ファイルシステム情報を反映した protobuf メッセージを構築します。
func (m *Koji) InitializeFromMessage(message *grpcv1.Koji) (*grpcv1.Koji, error) {
	// message が nil の場合はデフォルト初期化を行う
	if message == nil {
		return grpcv1.Koji_builder{}.Build(), nil
	}

	// dirPath を取得する
	dirPath := message.GetDirPath()

	// フォルダー名を取得
	dirName := core.GetBaseName(dirPath)

	// ファイル名から工事開始日の取得と日付除外文字列の取得
	start, dirNameRemovedDate, err := ParseTimestamp(dirName)
	if err != nil || dirNameRemovedDate == "" {
		return nil, errors.New("工事フォルダー名から工事開始日が取得できません error: " + err.Error())
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
		return nil, errors.New("工事フォルダー名から会社名及び現場名が得できません")
	}

	// Kojiメッセージの生成
	mes := grpcv1.Koji_builder{}.Build()

	// 各フィールドの設定
	mes.SetDirPath(dirPath)
	mes.SetStart(start.Timestamp)
	mes.SetCompanyName(companyName)
	mes.SetLocationName(locationName)

	m.EnsurePrEndFromStart(mes)

	// Id フィールドの設定
	newId := m.generateId(mes)
	mes.SetId(newId)

	return mes, nil
}

// UpdateMessage は source の値に基づいて target メッセージを更新します。
func (m *Koji) UpdateMessage(target *grpcv1.Koji, source *grpcv1.Koji) error {
	// 新しいパラメータを元に管理フォルダーパスを生成
	sourceStart := Timestamp{Timestamp: source.GetStart()}
	baseName, err := m.generateBaseName(
		sourceStart,
		source.GetCompanyName(),
		source.GetLocationName())
	if err != nil {
		return err
	}

	parentPath := filepath.Dir(target.GetDirPath())
	dirPath := filepath.Join(parentPath, baseName)

	// ディレクトリ名変更が必要な場合
	if dirPath != target.GetDirPath() {
		err := os.Rename(target.GetDirPath(), dirPath)
		if err != nil {
			return err
		}

		// フィールドの更新
		target.SetDirPath(dirPath)
		target.SetStart(source.GetStart())
		target.SetCompanyName(source.GetCompanyName())
		target.SetLocationName(source.GetLocationName())

		// pr_end が空の場合は start の値をコピー
		m.EnsurePrEndFromStart(target)

		// Id フィールドの更新
		newId := m.generateId(target)
		target.SetId(newId)
	}

	return nil
}

// generateId は dirPath から工事IDを生成します
func (m *Koji) generateId(message *grpcv1.Koji) string {
	dirPath := message.GetDirPath()
	basename := core.GetBaseName(dirPath)
	return core.BytesToId([]byte(basename))
}

// generateBaseName はパラメータをもとに工事フォルダー名変更します
func (m *Koji) generateBaseName(st Timestamp, cn string, loc string) (string, error) {
	// 開始日のフォーマット
	startText, err := st.FormatTime("2006-0102")
	if err != nil {
		return "", err
	}

	// フルパスの生成
	baseName := startText + " " + cn + " " + loc
	return baseName, nil
}

// EnsurePrEndFromStart は pr_end が空の場合に start の値をコピーします
func (m *Koji) EnsurePrEndFromStart(mes *grpcv1.Koji) bool {
	if mes == nil {
		return false
	}

	// pr_end が有効な場合は終了
	if mes.GetPrEnd() != nil && mes.GetPrEnd().IsValid() {
		return false
	}

	// start の無効な場合は終了
	if mes.GetStart() == nil || !mes.GetStart().IsValid() {
		return false
	}

	// start の値を pr_end にコピー
	mes.SetPrEnd(mes.GetStart())
	return true
}

// NewPersistModelKoji は指定された会社フォルダーから PersistModel[*Koji] を作成します。
// 新規実装ではこちらを使用してください。
func NewPersistModelKoji(dirPath string) (*core.PersistModel[*grpcv1.Koji, *Koji], error) {
	// PersistModel を作成
	pm, err := core.NewPersistModel(&Koji{}, "@koji.yaml")
	if err != nil {
		return nil, err
	}

	// 初期化
	request := grpcv1.Koji_builder{}.Build()
	request.SetDirPath(dirPath)
	err = pm.Initialize(request)
	if err != nil {
		return nil, err
	}

	return pm, nil
}

package models

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
	grpcv1 "web-api/gen/grpc/v1"
	"web-api/internal/core"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

// Koji core.PersistModel[*grpcv1.Koji] の拡張版です。
type Koji struct{}

// ParseFromDirPath は dirPath から工事開始日・会社名・現場名を取得
func (m *Koji) GenerateMessage(request protoreflect.Message) (protoreflect.ProtoMessage, error) {
	// mes が 空文字 の場合はデフォルト初期化を行う
	if request == nil {
		return grpcv1.Koji_builder{}.Build(), nil
	}
	// message の型アサーション
	req, ok := request.Interface().(*grpcv1.Koji)
	if !ok {
		return nil, errors.New("message の型アサーションに失敗しました")
	}
	// dirPath を取得する
	dirPath := req.GetDirPath()

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

	m.EnsureMfEndFromStart(mes)

	// Id フィールドの設定
	newId := m.GenerateId(mes.ProtoReflect())
	mes.SetId(newId)

	return mes, nil
}

// GenerateId は dirPath から工事IDを生成します
func (m *Koji) GenerateId(message protoreflect.Message) string {
	mes, ok := message.Interface().(*grpcv1.Koji)
	if !ok {
		return ""
	}
	dirPath := mes.GetDirPath()
	basename := core.GetBaseName(dirPath)
	return core.BytesToId([]byte(basename))
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

// UpdateMessage は情報を更新します
func (m *Koji) UpdateMessage(target protoreflect.Message, source protoreflect.Message) error {
	// source メッセージの型アサーション
	srcMes, ok := source.Interface().(*grpcv1.Koji)
	if !ok {
		return errors.New("source メッセージの型アサーションに失敗しました")
	}

	// 新しいパラメータを元に管理フォルダーパスを生成
	sourceStart := Timestamp{Timestamp: srcMes.GetStart()}
	baseName, err := m.generateBaseName(
		sourceStart,
		srcMes.GetCompanyName(),
		srcMes.GetLocationName())
	if err != nil {
		return err
	}

	// target メッセージの型アサーション
	mes, ok := target.Interface().(*grpcv1.Koji)
	if !ok {
		return errors.New("target メッセージの型アサーションに失敗しました")
	}

	parentPath := filepath.Dir(mes.GetDirPath())
	dirPath := filepath.Join(parentPath, baseName)

	// ディレクトリ名変更が必要な場合
	if dirPath != mes.GetDirPath() {
		err := os.Rename(mes.GetDirPath(), dirPath)
		if err != nil {
			return err
		}

		// フィールドの更新
		mes.SetDirPath(dirPath)
		mes.SetStart(srcMes.GetStart())
		mes.SetCompanyName(srcMes.GetCompanyName())
		mes.SetLocationName(srcMes.GetLocationName())

		// mf_end が空の場合は start の値をコピー
		m.EnsureMfEndFromStart(mes)

		// Id フィールドの更新
		newId := m.GenerateId(mes.ProtoReflect())
		mes.SetId(newId)
	}

	return nil
}

// generateBaseName はパラメータをもとに工事フォルダー名変更します
//
//	引数： st: 工事開始日
//	      cn: 会社名
//	      loc: 現場名
//	戻り値:
//
//		生成された工事フォルダーパス
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

// EnsureMfEndFromStart は mf_end が空の場合に start の値をコピーします
func (m *Koji) EnsureMfEndFromStart(mes *grpcv1.Koji) bool {

	if mes == nil {
		return false
	}

	// mf_end が有効な場合は終了
	if mes.GetMfEnd() != nil && mes.GetMfEnd().IsValid() {
		return false
	}

	// start の無効な場合は終了
	if mes.GetStart() == nil || !mes.GetStart().IsValid() {
		return false
	}

	// start の値を mf_end にコピー
	mes.SetMfEnd(mes.GetStart())
	return true
}

// NewPersistModelKoji は指定された会社フォルダーから PersistModel[*Koji] を作成します
func NewPersistModelKoji(dirPath string) (*core.PersistModel[*Koji], error) {
	// PersistModel を作成
	pm, err := core.NewPersistModel(&Koji{}, "koji.yaml")
	if err != nil {
		return nil, err
	}

	// 初期化
	request := grpcv1.Koji_builder{}.Build()
	request.SetDirPath(dirPath)
	err = pm.Initialize(request.ProtoReflect())
	if err != nil {
		return nil, err
	}

	//
	return pm, nil
}

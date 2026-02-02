package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// PersistModelの定義
// protobuf メッセージの特定のフィールド群を永続化するためのモデルです。
// 特定のフィールド群とは、フィールド名が "pr_" で始まるフィールドを指します。
// 一つの情報源はファイルシステムのパス名から取得されるのですが、それでは足りない場合があります。
type PersistModel[T Persistable] struct {
	Model           T
	Message         proto.Message
	persistFilename string
}

// Persistable は proto の pr_ フィールドをpersistファイルに保存できるモデルのインターフェースを定義します。
//   - protobuf メッセージを持っていることが前提となります。
type Persistable interface {
	// MessageFromDirPath は dirPath をもとにモデルの初期化を行います。
	// mes の情報をもとにメッセージの生成を行います。
	GenerateMessage(request proto.Message) (proto.Message, error)

	// UpdateMessage は target メッセージと source メッセージをもとに
	// 更新後のメッセージを生成します。
	UpdateMessage(target proto.Message, source proto.Message) error
}

// NewPersistModel は PersistModel インスタンスを作成します。
func NewPersistModel[T Persistable](model T, persistFileName string) (*PersistModel[T], error) {
	// モデルからデフォルトメッセージを取得
	mes, err := model.GenerateMessage(nil)
	if err != nil {
		return nil, err
	}

	// PersistModel インスタンスを作成
	return &PersistModel[T]{
		Model:           model,
		Message:         mes,
		persistFilename: persistFileName,
	}, nil
}

// Initialize は initArg をもとにモデルの初期化を行います。
// この関数が呼ばれたときに p.persistDirPath が設定されます。
func (p *PersistModel[T]) Initialize(request proto.Message) error {
	// dirPath から Messageを取得
	mes, err := p.Model.GenerateMessage(request)
	if err != nil {
		return err
	}

	// メッセージを設定
	p.Message = mes
	err = p.Load()
	if err != nil {
		return err
	}

	return nil
}

// GetPersistFilePath は p.persistFilename を p.message のフィールド "dir_path"から取得します。
func (p *PersistModel[T]) GetPersistFilePath() (string, error) {
	if p == nil || p.Message == nil {
		return "", errors.New("message is nil")
	}

	// dir_path フィールドの取得
	dirPath, err := GetFieldAs[string](p.Message, "dir_path")
	if err != nil {
		return "", err
	}

	filename := filepath.Base(dirPath)
	if filename == "" {
		return "", errors.New("invalid dir_path value")
	}

	return filepath.Join(dirPath, p.persistFilename), nil
}

// Load は Persist ファイルから永続化データのみを読み込みます。
// ファイル形式は YAML です。
func (p *PersistModel[T]) Load() error {

	// YAMLファイルからテキストデータを読み込む
	persistFilePath, err := p.GetPersistFilePath()
	if err != nil {
		return err
	}
	text, err := os.ReadFile(persistFilePath)
	if err != nil {
		// ファイルが存在しない場合は新規作成
		return p.Save()
	}

	// YAMLファイルデータをJSONマップデータに変換
	jsonmap := &map[string]any{}
	err = yaml.Unmarshal(text, jsonmap)
	if len(*jsonmap) == 0 || err != nil {
		return p.Save()
	}

	// JSONマップデータから Manifest データを取り込む
	err = p.ImportJson(jsonmap)
	if err != nil {
		return p.Save()
	}
	return nil
}

// Save は protobuf メッセージの 接頭語が "pr_" で始まるデータを Persist ファイルに保存します。
// ファイル形式は YAML です。
func (p *PersistModel[T]) Save() error {
	// JSONマップの取得
	jsonmap, err := p.ExportJson()
	if err != nil {
		return err
	}

	// JSONマップをYAMLデータに変換
	yamlBytes, err := yaml.Marshal(jsonmap)
	if err != nil {
		return err
	}

	// ファイルに書き込み
	persistFilePath, err := p.GetPersistFilePath()
	if err != nil {
		return err
	}
	return os.WriteFile(persistFilePath, yamlBytes, 0644)
}

// UpdatePersistFields は Persist データを更新します。
//
// source: PersistModel[T] - 更新元の PersistModel インスタンス
func (p *PersistModel[T]) UpdatePersistFields(source *PersistModel[T]) error {
	// 引数チェック
	if source == nil || source.Message == nil {
		return errors.New("Source PersistModel src is nil")
	}

	// pのTとsourceのTが同じ型であることを確認
	if reflect.TypeOf(p.Model) != reflect.TypeOf(source.Model) {
		return errors.New("model type mismatch")
	}

	// Persist フィールドのみを更新
	fields := p.Message.ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		v := source.Message.ProtoReflect().Get(f)
		name := string(f.Name())
		if !strings.HasPrefix(name, "pr_") {
			continue
		}
		p.Message.ProtoReflect().Set(f, v)
	}
	return nil
}

func (p *PersistModel[T]) Update(source *PersistModel[T]) error {

	// source が nil の場合は m.dirPath から再初期化を行う
	if source == nil {
		// メッセージの再生成
		mes, err := p.Model.GenerateMessage(p.Message)
		if err != nil {
			return err
		}
		p.Message = mes

		// Persist データのロード
		return p.Load()
	}

	// source.Message データをもとに新たなメッセージを生成
	err := p.Model.UpdateMessage(p.Message, source.Message)
	if err != nil {
		return err
	}

	// Persist データのロード
	return p.Load()
}

// ExportJson は Persist フィールド値をJSONに変換します
func (p *PersistModel[T]) ExportJson() (*map[string]any, error) {
	// camelCase キーで JSON にマーシャル
	jsonbytes, err := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}.Marshal(p.Message)
	if err != nil {
		return nil, err
	}

	// pr_ フィールドのみを抽出した JSON マップを作成
	jsonmap := &map[string]any{}
	json.Unmarshal(jsonbytes, jsonmap)
	for k := range *jsonmap {
		if !strings.HasPrefix(k, "pr_") {
			delete(*jsonmap, k)
		}
	}

	// タイムスタンプフィールドを日本時間（JST）に変換
	p.convertTimestampsToJST(jsonmap)

	return jsonmap, nil
}

// convertTimestampsToJST はマップ内のタイムスタンプを日本時間のフォーマットに変換します
func (p *PersistModel[T]) convertTimestampsToJST(jsonmap *map[string]any) {
	if p == nil || p.Message == nil {
		return
	}

	ref := p.Message.ProtoReflect()
	if ref == nil {
		return
	}
	fields := ref.Descriptor().Fields()
	jst := time.FixedZone("JST", 9*60*60)

	// 各フィールドを確認
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		fieldName := string(f.Name())

		// pr_ で始まるフィールドのみ処理
		if !strings.HasPrefix(fieldName, "pr_") {
			continue
		}

		// Timestamp 型のフィールドのみ処理
		if f.Message() != nil && f.Message().FullName() == "google.protobuf.Timestamp" {
			if ref.Has(f) {
				value := ref.Get(f)
				if ts, ok := value.Message().Interface().(*timestamppb.Timestamp); ok && ts.IsValid() {
					// JST で RFC3339 フォーマットに変換
					jstTime := ts.AsTime().In(jst)
					(*jsonmap)[fieldName] = jstTime.Format(time.RFC3339)
				}
			}
		}
	}
}

// ImportJson はJSONマップを Persist フィールドに設定します
func (p *PersistModel[T]) ImportJson(jsonmap *map[string]any) error {

	// タイムスタンプ文字列をUTC形式に正規化
	p.normalizeTimestampsToUTC(jsonmap)

	// JSONマップをバイトデータに変換
	bytes, err := json.Marshal(*jsonmap)
	if err != nil {
		return err
	}

	// p.Message の ProtoReflect を取得
	msgRef := p.Message.ProtoReflect()

	// 一時的な空のメッセージを作成してアンマーシャル
	tempMsg := msgRef.Type().New()

	opts := protojson.UnmarshalOptions{AllowPartial: true}
	if err := opts.Unmarshal(bytes, tempMsg.Interface()); err != nil {
		return err
	}

	// pr_ フィールドのみを元のメッセージにコピー
	fields := msgRef.Descriptor().Fields()
	tempRef := tempMsg
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		if !strings.HasPrefix(string(f.Name()), "pr_") {
			continue
		}
		if tempRef.Has(f) {
			msgRef.Set(f, tempRef.Get(f))
		}
	}

	return nil
}

// normalizeTimestampsToUTC はマップ内のタイムスタンプ文字列をUTC形式に正規化します
func (p *PersistModel[T]) normalizeTimestampsToUTC(jsonmap *map[string]any) {
	if jsonmap == nil {
		return
	}

	// 各フィールドを確認
	for k, v := range *jsonmap {
		// pr_ で始まるフィールドのみ処理
		if !strings.HasPrefix(k, "pr_") {
			continue
		}

		// 文字列型の値のみ処理
		if strVal, ok := v.(string); ok {
			// RFC3339形式のタイムスタンプとしてパース
			if t, err := time.Parse(time.RFC3339, strVal); err == nil {
				// UTC形式の文字列に変換
				(*jsonmap)[k] = t.UTC().Format(time.RFC3339)
			}
		}
	}
}

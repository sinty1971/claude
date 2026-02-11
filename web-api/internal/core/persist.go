package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// Persistable は型安全な Persistable インターフェースです。
// Generics を使用することで、コンパイル時の型チェックが効き、
// 実行時の型アサーションエラーを防ぐことができます。
//
// 新規実装ではこちらを使用してください。既存の Persistable からの移行も推奨します。
type Persistable[M proto.Message] interface {
	// InitializeFromDirPath は message メッセージを元に、ファイルシステム情報を反映した protobuf メッセージを構築します。
	// message が nil の場合は、デフォルト初期化されたメッセージを返します。
	// message が非 nil の場合は、message.DirPath などのファイルシステム情報を解析し、
	// 対応するドメインモデル（ID, Name, Category など）を設定したメッセージを返します。
	//
	// 型パラメータ M により、型安全性が保証されます。
	InitializeFromMessage(message M) (M, error)

	// UpdateMessage は source の値に基づいて target メッセージを更新します。
	// この処理には、メッセージフィールドの更新に加えて、必要に応じてファイルシステム操作
	// （ディレクトリ名の変更など）も含まれる場合があります。
	//
	// 例: 会社名が変更された場合、"1 旧会社名" → "1 新会社名" のようにディレクトリをリネームし、
	// target.DirPath や target.Id などのフィールドを更新します。
	//
	// 型パラメータ M により、型安全性が保証されます。
	UpdateMessage(target M, source M) error
}

// PersistModel は型安全な PersistModel です。
// Generics を使用することで、コンパイル時の型チェックが効き、
// 実行時の型アサーションエラーを防ぐことができます。
//
// 新規実装ではこちらを使用してください。既存の PersistModel からの移行も推奨します。
type PersistModel[M proto.Message, T Persistable[M]] struct {
	Model           T
	Message         M
	persistFilename string
}

// NewPersistModel は TypedPersistModel インスタンスを作成します。
//
//	model: TypedPersistable モデルインスタンス
//	persistFileName: 永続化ファイル名
func NewPersistModel[M proto.Message, T Persistable[M]](model T, persistFileName string) (*PersistModel[M, T], error) {
	// モデルからデフォルトメッセージを取得
	var nilMsg M
	mes, err := model.InitializeFromMessage(nilMsg)
	if err != nil {
		return nil, err
	}

	// TypedPersistModel インスタンスを作成
	return &PersistModel[M, T]{
		Model:           model,
		Message:         mes,
		persistFilename: persistFileName,
	}, nil
}

// Initialize は message をもとにモデルの初期化を行います。
func (p *PersistModel[M, T]) Initialize(message M) error {
	// メッセージを取得
	mes, err := p.Model.InitializeFromMessage(message)
	if err != nil {
		return err
	}

	// メッセージを設定
	p.Message = mes

	// Persist データのロード
	err = p.Load()
	if err != nil {
		return err
	}

	return nil
}

// Load は Persist ファイルから永続化データのみを読み込みます。
// ファイル形式は YAML です。
func (p *PersistModel[M, T]) Load() error {
	// YAMLファイルからテキストデータを読み込む
	persistFilePath, err := p.GetPersistFilePath()
	if err != nil {
		return err
	}

	// ファイル読み込み
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
func (p *PersistModel[M, T]) Save() error {
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

// GetPersistFilePath は p.persistFilename を p.message のフィールド "dir_path"から取得します。
func (p *PersistModel[M, T]) GetPersistFilePath() (string, error) {
	if p == nil {
		return "", errors.New("ポインタレシーバがnilです")
	}

	// M を proto.Message にキャスト
	msg, ok := any(p.Message).(proto.Message)
	if !ok {
		return "", errors.New("Message が proto.Message にキャストできません")
	}

	// dir_path フィールドの取得
	dirPath, err := GetFieldAs[string](msg, "dir_path")
	if err != nil {
		return "", err
	}

	filename := filepath.Base(dirPath)
	if filename == "" {
		return "", errors.New("dir_path の値が無効です")
	}

	return filepath.Join(dirPath, p.persistFilename), nil
}

// Update は source データをもとに更新を行います。
func (p *PersistModel[M, T]) Update(source *PersistModel[M, T]) error {
	// source が nil の場合は p.Message から再初期化を行う
	if source == nil {
		// メッセージの再生成
		mes, err := p.Model.InitializeFromMessage(p.Message)
		if err != nil {
			return err
		}
		p.Message = mes

		// Persist データのロード
		return p.Load()
	}

	// source.Message データをもとに新たなメッセージを更新
	err := p.Model.UpdateMessage(p.Message, source.Message)
	if err != nil {
		return err
	}

	// Persist データのロード
	return p.Load()
}

// UpdatePersistFields は Persist データを更新します。
func (p *PersistModel[M, T]) UpdatePersistFields(source *PersistModel[M, T]) error {
	// 引数チェック
	if source == nil {
		return errors.New("Source TypedPersistModel is nil")
	}

	// M を proto.Message にキャスト
	targetMsg, ok1 := any(p.Message).(proto.Message)
	sourceMsg, ok2 := any(source.Message).(proto.Message)
	if !ok1 || !ok2 {
		return errors.New("Message を proto.Message にキャストできません")
	}

	// Persist フィールドのみを更新
	fields := targetMsg.ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		v := sourceMsg.ProtoReflect().Get(f)
		name := string(f.Name())
		if !strings.HasPrefix(name, "pr_") {
			continue
		}
		targetMsg.ProtoReflect().Set(f, v)
	}
	return nil
}

// ExportJson は Persist フィールド値をJSONに変換します
func (p *PersistModel[M, T]) ExportJson() (*map[string]any, error) {
	// M を proto.Message にキャスト
	msg, ok := any(p.Message).(proto.Message)
	if !ok {
		return nil, errors.New("Message が proto.Message にキャストできません")
	}

	// camelCase キーで JSON にマーシャル
	jsonbytes, err := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}.Marshal(msg)
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
	p.convertTimestampsToJST(jsonmap, msg)

	return jsonmap, nil
}

// convertTimestampsToJST はマップ内のタイムスタンプを日本時間のフォーマットに変換します
func (p *PersistModel[M, T]) convertTimestampsToJST(jsonmap *map[string]any, msg proto.Message) {
	if msg == nil {
		return
	}

	ref := msg.ProtoReflect()
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
func (p *PersistModel[M, T]) ImportJson(jsonmap *map[string]any) error {
	// タイムスタンプ文字列をUTC形式に正規化
	p.normalizeTimestampsToUTC(jsonmap)

	// JSONマップをバイトデータに変換
	bytes, err := json.Marshal(*jsonmap)
	if err != nil {
		return err
	}

	// M を proto.Message にキャスト
	msg, ok := any(p.Message).(proto.Message)
	if !ok {
		return errors.New("Message が proto.Message にキャストできません")
	}

	// p.Message の ProtoReflect を取得
	msgRef := msg.ProtoReflect()

	// 一時的な空のメッセージを作成してアンマーシャル
	tempMsg := msgRef.Type().New()

	// pr_ フィールドのみを含む一時メッセージにアンマーシャル
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
func (p *PersistModel[M, T]) normalizeTimestampsToUTC(jsonmap *map[string]any) {
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

package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

// ManifestProviderの定義
// 一つの情報源はファイルシステムのパス名から取得されるのですが、それでは足りない場合があります。
// 例えば、同じ会社に対して複数の永続化データを持ちたい場合などです。
// そこで、ManifestProvider構造体を用いて、各モデルに対して追加の永続化データを管理します。
// そのManifestProviderデータは対象フォルダ内の特定のYAMLファイルに保存されます。
// ManifestProvider は永続化設定を保持します。
// モデルへの参照は保持せず、循環参照を避けています。
type ManifestProvider struct {
	// Manifestable インターフェースを実装するモデルへの参照
	Manifestable

	// マニフェストファイル名
	ManifestFileName string
}

// Manifestable は proto の mf_ フィールドをmanifestファイルに保存できるモデルのインターフェースを定義します。
//   - protobuf メッセージを持っていることが前提となります。
type Manifestable interface {
	// InitializeManifestProvider は ManifestProvider を初期化します。
	InitializeManifestProvider()

	// GetProtoMessage はモデルの protobuf メッセージを取得します。
	ProtoReflect() protoreflect.Message

	// GetDirPath はモデルのディレクトリパスを取得します。
	GetDirPath() string

	// GetId はモデルのIDを取得します。
	GetId() string

	// Save はマニフェストを保存します。
	Save() error

	// Load はマニフェストを読み込みます。
	Load() error
}

// GetMessageFullName はモデルの protobuf メッセージの完全修飾名を取得します。
func (p *ManifestProvider) GetMessageFullName() string {
	if p == nil || p.Manifestable == nil || p.Manifestable.ProtoReflect() == nil {
		return ""
	}
	return string(p.Manifestable.ProtoReflect().Descriptor().FullName())
}

func (p *ManifestProvider) getManifestPath() string {
	return filepath.Join(p.Manifestable.GetDirPath(), p.ManifestFileName)
}

// Load は Manifest ファイルから永続化データのみを読み込みます。
// ファイル形式は YAML です。
func (p *ManifestProvider) Load() error {

	// YAMLファイルからテキストデータを読み込む
	text, err := os.ReadFile(p.getManifestPath())
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

// Save はデータを Manifest ファイルに保存します。
// ファイル形式は YAML です。
func (p *ManifestProvider) Save() error {
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
	return os.WriteFile(p.getManifestPath(), yamlBytes, 0644)
}

// Update は Manifest データを更新します。
//
// source: ManifestProvider
func (p *ManifestProvider) UpdateManifest(source *ManifestProvider) error {
	// 引数チェック
	if source == nil || source.Manifestable == nil {
		return errors.New("Source Manifestable src is nil")
	}

	// モデル名チェック
	if p.GetMessageFullName() != source.GetMessageFullName() {
		return errors.New("MessageFullName mismatch")
	}

	// Manifest フィールドのみを更新
	targetRef := p.Manifestable.ProtoReflect()
	fields := targetRef.Descriptor().Fields()
	srcRef := source.Manifestable.ProtoReflect()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		v := srcRef.Get(f)
		name := string(f.Name())
		if !strings.HasPrefix(name, "mf_") {
			continue
		}
		targetRef.Set(f, v)
	}
	return nil
}

// ExportJson は Manifest フィールド値をJSONに変換します
func (p *ManifestProvider) ExportJson() (*map[string]any, error) {
	// camelCase キーで JSON にマーシャル
	jsonbytes, err := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}.Marshal(p.Manifestable.ProtoReflect().Interface())
	if err != nil {
		return nil, err
	}

	// mf_ フィールドのみを抽出した JSON マップを作成
	jsonmap := &map[string]any{}
	json.Unmarshal(jsonbytes, jsonmap)
	for k := range *jsonmap {
		if !strings.HasPrefix(k, "mf_") {
			delete(*jsonmap, k)
		}
	}

	// タイムスタンプフィールドを日本時間（JST）に変換
	p.convertTimestampsToJST(jsonmap)

	return jsonmap, nil
}

// convertTimestampsToJST はマップ内のタイムスタンプを日本時間のフォーマットに変換します
func (p *ManifestProvider) convertTimestampsToJST(jsonmap *map[string]any) {
	if p == nil || p.Manifestable == nil {
		return
	}

	ref := p.Manifestable.ProtoReflect()
	if ref == nil {
		return
	}
	fields := ref.Descriptor().Fields()
	jst := time.FixedZone("JST", 9*60*60)

	// 各フィールドを確認
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		fieldName := string(f.Name())

		// mf_ で始まるフィールドのみ処理
		if !strings.HasPrefix(fieldName, "mf_") {
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

// ImportJson はJSONマップを Manifest フィールドに設定します
func (p *ManifestProvider) ImportJson(jsonmap *map[string]any) error {

	// タイムスタンプ文字列をUTC形式に正規化
	p.normalizeTimestampsToUTC(jsonmap)

	// JSONマップをバイトデータに変換
	bytes, err := json.Marshal(*jsonmap)
	if err != nil {
		return err
	}

	// 一時的な空のメッセージを作成してアンマーシャル
	targetRef := p.Manifestable.ProtoReflect()
	tempMsg := targetRef.Type().New()

	opts := protojson.UnmarshalOptions{AllowPartial: true}
	if err := opts.Unmarshal(bytes, tempMsg.Interface()); err != nil {
		return err
	}

	// mf_ フィールドのみを元のメッセージにコピー
	fields := targetRef.Descriptor().Fields()
	tempRef := tempMsg
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		if !strings.HasPrefix(string(f.Name()), "mf_") {
			continue
		}
		if tempRef.Has(f) {
			targetRef.Set(f, tempRef.Get(f))
		}
	}

	return nil
}

// normalizeTimestampsToUTC はマップ内のタイムスタンプ文字列をUTC形式に正規化します
func (p *ManifestProvider) normalizeTimestampsToUTC(jsonmap *map[string]any) {
	if jsonmap == nil {
		return
	}

	// 各フィールドを確認
	for k, v := range *jsonmap {
		// mf_ で始まるフィールドのみ処理
		if !strings.HasPrefix(k, "mf_") {
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

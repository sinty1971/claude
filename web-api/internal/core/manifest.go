package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
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
}

// Manifestable はmanifestフィールドをmanifestファイルに持つモデルのインターフェースを定義します。
//   - protobuf メッセージを持っていることが前提となります。
type Manifestable interface {
	// GetManifestDirectory は永続化ファイルを保存するフルパスを取得します。
	//  - proto メッセージ内の mf_folder フィールドを返す実装が一般的です。
	GetManifestDirectory() string

	// GetManifestMessage はモデルの protobuf メッセージを取得します。
	GetManifestMessage() proto.Message
}

// NewManifestProvider は ManifestProvider のインスタンスを作成します。
func NewManifestProvider(target Manifestable) *ManifestProvider {
	// インスタンス作成（モデルへの参照は保持しない）
	return &ManifestProvider{
		Manifestable: target,
	}
}

// GetMessageFullName はモデルの protobuf メッセージの完全修飾名を取得します。
func (p *ManifestProvider) GetMessageFullName() string {
	if p == nil || p.Manifestable.GetManifestMessage() == nil {
		return ""
	}
	return string(p.Manifestable.GetManifestMessage().ProtoReflect().Descriptor().FullName())
}

func (p *ManifestProvider) getManifestPath() string {
	return filepath.Join(p.GetManifestDirectory(), "@manifest.yaml")
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
func (p *ManifestProvider) Update(source *ManifestProvider) error {
	// 引数チェック
	if source == nil || source.Manifestable == nil {
		return errors.New("Source Manifestable src is nil")
	}

	// モデル名チェック
	if p.GetMessageFullName() != source.GetMessageFullName() {
		return errors.New("MessageFullName mismatch")
	}

	// Manifest フィールドのみを更新
	targetRef := p.GetManifestMessage().ProtoReflect()
	fields := targetRef.Descriptor().Fields()
	srcRef := source.GetManifestMessage().ProtoReflect()
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
		UseProtoNames:     true,
		EmitUnpopulated:   false,
		EmitDefaultValues: true,
	}.Marshal(p.GetManifestMessage())
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

	return jsonmap, nil
}

// ImportJson はJSONマップを Manifest フィールドに設定します
func (p *ManifestProvider) ImportJson(jsonmap *map[string]any) error {

	// JSONマップをバイトデータに変換
	bytes, err := json.Marshal(*jsonmap)
	if err != nil {
		return err
	}

	// 代入先メッセージの取得
	targetRef := p.GetManifestMessage().ProtoReflect()
	fields := targetRef.Descriptor().Fields()

	// JSONデータをアンマーシャルし、一時メッセージに格納
	tempMsg := dynamicpb.NewMessage(targetRef.Descriptor())
	opts := protojson.UnmarshalOptions{AllowPartial: true}
	if err := opts.Unmarshal(bytes, tempMsg); err != nil {
		return err
	}

	// Manifest フィールドのみを元のメッセージにコピー
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		if !strings.HasPrefix(string(f.Name()), "mf_") {
			continue
		}
		targetRef.Set(f, tempMsg.Get(f))
	}
	return nil
}

package core

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Repository は永続化を自動管理するジェネリックリポジトリ
type Repository[T Persistable] struct {
	cache    map[string]PersistModel[T]
	mu       sync.RWMutex
	autoSave bool
}

// NewRepository は新しいRepositoryを作成
func NewRepository[T Persistable](autoSave bool) *Repository[T] {
	return &Repository[T]{
		cache:    make(map[string]PersistModel[T]),
		autoSave: autoSave,
	}
}

// Get はIDでアイテムを取得
func (r *Repository[T]) Get(id string) (PersistModel[T], bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, exists := r.cache[id]
	return m, exists
}

// Set はアイテムを設定し、autoSaveが有効なら自動保存
func (r *Repository[T]) Set(item PersistModel[T]) error {
	// ロックして設定
	r.mu.Lock()
	defer r.mu.Unlock()

	id, err := GetFieldAs[string](item.Message, "id")
	if err != nil {
		return fmt.Errorf("failed to get id: %w", err)
	}
	r.cache[id] = item

	if r.autoSave {
		if err := item.Save(); err != nil {
			return fmt.Errorf("auto-save failed for id=%s: %w", id, err)
		}
	}

	return nil
}

// Update は既存アイテムを更新し、autoSaveが有効なら自動保存
func (r *Repository[T]) Update(targetId string, source proto.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	target, exists := r.cache[targetId]
	if !exists {
		return fmt.Errorf("item not found: id=%s", targetId)
	}

	if err := target.Model.UpdateMessage(target.Message, source); err != nil {
		return fmt.Errorf("update function failed: %w", err)
	}

	// 更新後のIDを取得
	updatedId, err := GetFieldAs[string](target.Message, "id")
	if err != nil {
		return fmt.Errorf("failed to get updated id: %w", err)
	}

	// 更新後のIDが変わっていたらキャッシュキー taretId を削除
	if updatedId != targetId {
		delete(r.cache, targetId)
	}

	r.cache[updatedId] = target

	if r.autoSave {
		if err := target.Save(); err != nil {
			return fmt.Errorf("auto-save failed after update for id=%s: %w", updatedId, err)
		}
	}

	return nil
}

// Delete はアイテムを削除
func (r *Repository[T]) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.cache, id)
	return nil
}

// GetAllAsMessage は全アイテムのメッセージを取得
func (r *Repository[T]) GetAllAsMessage() []proto.Message {
	r.mu.RLock()
	defer r.mu.RUnlock()

	messages := make([]proto.Message, 0, len(r.cache))
	for _, item := range r.cache {
		messages = append(messages, item.Message)
	}
	return messages
}

// Count はアイテム数を返す
func (r *Repository[T]) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cache)
}

// Clear は全アイテムをクリア
func (r *Repository[T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]PersistModel[T])
}

// SetAutoSave は自動保存の有効/無効を切り替え
func (r *Repository[T]) SetAutoSave(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoSave = enabled
}

// SaveAll は全アイテムを強制保存
func (r *Repository[T]) SaveAll() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, item := range r.cache {
		if err := item.Save(); err != nil {
			return fmt.Errorf("failed to save item id=%s: %w", id, err)
		}
	}
	return nil
}

// AssertProtoAs は簡易コンバータ生成ヘルパーです。型アサーションを行い、失敗したらエラーを返します。
func AssertProtoAs[R any](m proto.Message) (R, error) {
	var zero R
	if m == nil {
		return zero, fmt.Errorf("message is nil")
	}
	iface := m
	if v, ok := iface.(R); ok {
		return v, nil
	}
	return zero, fmt.Errorf("cannot assert message to %T", zero)
}

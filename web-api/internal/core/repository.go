package core

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
)

// TypedRepository は型安全な永続化リポジトリです。
// 新規実装ではこちらを使用してください。
type TypedRepository[M proto.Message, T Persistable[M]] struct {
	cache    map[string]PersistModel[M, T]
	mu       sync.RWMutex
	autoSave bool
}

// NewTypedRepository は新しい TypedRepository を作成します。
func NewTypedRepository[M proto.Message, T Persistable[M]](autoSave bool) *TypedRepository[M, T] {
	return &TypedRepository[M, T]{
		cache:    make(map[string]PersistModel[M, T]),
		autoSave: autoSave,
	}
}

// Get はIDでアイテムを取得します。
func (r *TypedRepository[M, T]) Get(id string) (PersistModel[M, T], bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, exists := r.cache[id]
	return m, exists
}

// Set はアイテムを設定し、autoSaveが有効なら自動保存します。
func (r *TypedRepository[M, T]) Set(item PersistModel[M, T]) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// M を proto.Message にキャスト
	msg, ok := any(item.Message).(proto.Message)
	if !ok {
		return fmt.Errorf("failed to cast Message to proto.Message")
	}

	id, err := GetFieldAs[string](msg, "id")
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

// Update は既存アイテムを更新し、autoSaveが有効なら自動保存します。
func (r *TypedRepository[M, T]) Update(targetId string, source M) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	target, exists := r.cache[targetId]
	if !exists {
		return fmt.Errorf("item not found: id=%s", targetId)
	}

	if err := target.Model.UpdateMessage(target.Message, source); err != nil {
		return fmt.Errorf("update function failed: %w", err)
	}

	// M を proto.Message にキャスト
	msg, ok := any(target.Message).(proto.Message)
	if !ok {
		return fmt.Errorf("failed to cast Message to proto.Message")
	}

	// 更新後のIDを取得
	updatedId, err := GetFieldAs[string](msg, "id")
	if err != nil {
		return fmt.Errorf("failed to get updated id: %w", err)
	}

	// 更新後のIDが変わっていたらキャッシュキーを更新
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

// Delete はアイテムを削除します。
func (r *TypedRepository[M, T]) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.cache, id)
	return nil
}

// GetAllAsMessage は全アイテムのメッセージを取得します。
func (r *TypedRepository[M, T]) GetAllAsMessage() []M {
	r.mu.RLock()
	defer r.mu.RUnlock()

	messages := make([]M, 0, len(r.cache))
	for _, item := range r.cache {
		messages = append(messages, item.Message)
	}
	return messages
}

// Count はアイテム数を返します。
func (r *TypedRepository[M, T]) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cache)
}

// Clear は全アイテムをクリアします。
func (r *TypedRepository[M, T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]PersistModel[M, T])
}

// SetAutoSave は自動保存の有効/無効を切り替えます。
func (r *TypedRepository[M, T]) SetAutoSave(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoSave = enabled
}

// SaveAll は全アイテムを強制保存します。
func (r *TypedRepository[M, T]) SaveAll() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, item := range r.cache {
		if err := item.Save(); err != nil {
			return fmt.Errorf("failed to save item id=%s: %w", id, err)
		}
	}
	return nil
}

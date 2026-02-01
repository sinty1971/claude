package core

import (
	"fmt"
	"sync"
)

// Repository は永続化を自動管理するジェネリックリポジトリ
type Repository[T Persistable1] struct {
	cache    map[string]T
	mu       sync.RWMutex
	autoSave bool
}

// NewRepository は新しいRepositoryを作成
func NewRepository[T Persistable1](autoSave bool) *Repository[T] {
	return &Repository[T]{
		cache:    make(map[string]T),
		autoSave: autoSave,
	}
}

// Get はIDでアイテムを取得
func (r *Repository[T]) Get(id string) (T, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, exists := r.cache[id]
	return item, exists
}

// Set はアイテムを設定し、autoSaveが有効なら自動保存
func (r *Repository[T]) Set(id string, item T) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[id] = item

	if r.autoSave {
		if err := item.Save(); err != nil {
			return fmt.Errorf("auto-save failed for id=%s: %w", id, err)
		}
	}

	return nil
}

// Update は既存アイテムを更新し、autoSaveが有効なら自動保存
func (r *Repository[T]) Update(id string, updateFn func(T) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.cache[id]
	if !exists {
		return fmt.Errorf("item not found: id=%s", id)
	}

	if err := updateFn(item); err != nil {
		return fmt.Errorf("update function failed: %w", err)
	}

	r.cache[id] = item

	if r.autoSave {
		if err := item.Save(); err != nil {
			return fmt.Errorf("auto-save failed after update for id=%s: %w", id, err)
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

// GetAll は全アイテムを取得
func (r *Repository[T]) GetAll() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]T, 0, len(r.cache))
	for _, item := range r.cache {
		items = append(items, item)
	}
	return items
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
	r.cache = make(map[string]T)
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

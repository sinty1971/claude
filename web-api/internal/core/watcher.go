package core

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Watcher は指定されたディレクトリツリーを監視し、ファイルシステムの変更イベントを通知します
type Watcher struct {
	// watcher は fsnotify の監視オブジェクト
	watcher *fsnotify.Watcher

	// rootPath は監視対象のルートディレクトリ
	rootPath string

	// maxDepth は監視するディレクトリの最大深度
	maxDepth int

	// events は監視イベントを通知するチャネル
	events chan fsnotify.Event

	// watchedDirs は監視登録済みディレクトリの集合
	watchedDirs map[string]struct{}

	// done は監視ループを終了するためのチャネル
	done chan struct{}

	// errors はエラーを通知するチャネル
	errors chan error
}

// NewWatcher は新しい Watcher インスタンスを作成します
func NewWatcher() (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher:     fsWatcher,
		watchedDirs: make(map[string]struct{}),
		events:      make(chan fsnotify.Event),
		errors:      make(chan error),
		done:        make(chan struct{}),
	}, nil
}

// Start は監視を開始します
//
//	rootPath は監視対象のルートディレクトリ、maxDepth は監視するディレクトリの最大深度を指定します
//	maxDepth が 0 の場合、rootPath のみが監視されます
func (w *Watcher) Start(rootPath string, maxDepth int) error {
	w.rootPath = rootPath
	w.maxDepth = maxDepth

	if err := w.addWatchersRecursively(w.rootPath, 0); err != nil {
		return err
	}

	go w.loop()
	return nil
}

// Close は監視を停止し、リソースを解放します
func (w *Watcher) Close() error {
	close(w.done)
	return w.watcher.Close()
}

// Events はイベントチャネルを返します
func (w *Watcher) Events() <-chan fsnotify.Event {
	return w.events
}

// Errors はエラーチャネルを返します
func (w *Watcher) Errors() <-chan error {
	return w.errors
}

// loop は fsnotify からのイベントを監視し、内部処理と外部への通知を行います
func (w *Watcher) loop() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleInternalEvent(event)

			// 外部にイベントを転送
			select {
			case w.events <- event:
			case <-w.done:
				return
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			select {
			case w.errors <- err:
			case <-w.done:
				return
			}
		case <-w.done:
			return
		}
	}
}

// handleInternalEvent は内部的なイベント処理（ディレクトリ監視の追加・削除など）を行います
func (w *Watcher) handleInternalEvent(event fsnotify.Event) {
	if event.Name == "" {
		return
	}

	// ディレクトリ作成イベントの場合は監視対象を拡張
	if event.Op&fsnotify.Create != 0 {
		if err := w.addWatchIfDirectory(event.Name); err != nil {
			log.Printf("Watcher: Failed to expand watcher for %s: %v", event.Name, err)
		}
	}

	// ディレクトリ削除・移動イベントの場合は監視対象を解除
	if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		w.unregisterWatcherTree(event.Name)
	}
}

// addWatchIfDirectory は指定されたパスがディレクトリであれば監視対象に追加します
func (w *Watcher) addWatchIfDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil
	}

	depth, ok := w.relativeDepth(path)
	if !ok || depth > w.maxDepth {
		return nil
	}

	return w.addWatchersRecursively(path, depth)
}

// addWatchersRecursively は指定されたディレクトリとそのサブディレクトリを再帰的に監視対象に追加します
func (w *Watcher) addWatchersRecursively(dir string, depth int) error {
	if depth > w.maxDepth {
		return nil
	}

	cleanDir := filepath.Clean(dir)
	if err := w.registerWatcher(cleanDir); err != nil {
		return err
	}

	if depth == w.maxDepth {
		return nil
	}

	entries, err := os.ReadDir(cleanDir)
	if err != nil {
		// ディレクトリが読めない場合はログを出して続行（削除された場合など）
		log.Printf("Watcher: Failed to read directory %s: %v", cleanDir, err)
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if err := w.addWatchersRecursively(filepath.Join(cleanDir, entry.Name()), depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}

// registerWatcher は指定されたディレクトリを fsnotify の監視対象に登録します
func (w *Watcher) registerWatcher(target string) error {
	if _, watched := w.watchedDirs[target]; watched {
		return nil
	}

	if err := w.watcher.Add(target); err != nil {
		return err
	}
	w.watchedDirs[target] = struct{}{}

	return nil
}

// unregisterWatcherTree は指定されたディレクトリとその配下の監視を解除します
func (w *Watcher) unregisterWatcherTree(target string) {
	cleanPath := filepath.Clean(target)
	if _, ok := w.relativeDepth(cleanPath); !ok {
		return
	}

	targets := make([]string, 0, len(w.watchedDirs))
	prefix := cleanPath + string(os.PathSeparator)
	for dir := range w.watchedDirs {
		if dir == cleanPath || strings.HasPrefix(dir, prefix) {
			targets = append(targets, dir)
		}
	}

	for _, dir := range targets {
		if err := w.watcher.Remove(dir); err != nil {
			// 既に削除されている場合などはエラーになる可能性があるが、ログ出力のみとする
			log.Printf("Watcher: Failed to remove watcher for %s: %v", dir, err)
		}
		delete(w.watchedDirs, dir)
	}
}

// relativeDepth はルートパスからの相対的な深さを計算します
func (w *Watcher) relativeDepth(target string) (int, bool) {
	rel, err := filepath.Rel(w.rootPath, target)
	if err != nil {
		return -1, false
	}

	slashed := filepath.ToSlash(rel)
	if slashed == "." || slashed == "" {
		return 0, true
	}

	if slashed == ".." || strings.HasPrefix(slashed, "../") {
		return -1, false
	}

	return strings.Count(slashed, "/") + 1, true
}

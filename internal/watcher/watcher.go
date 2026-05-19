package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	// debounceInterval is the quiet period before a file change is acted upon.
	// Prevents duplicate processing when editors perform atomic saves.
	debounceInterval = 300 * time.Millisecond
)

// EventHandler is called after debouncing settles for a given file.
type EventHandler func(path string)

// Watcher wraps fsnotify with debouncing, recursive directory watching,
// and event routing based on file type.
type Watcher struct {
	ctx             context.Context
	root            string
	fsWatcher       *fsnotify.Watcher
	onSchemaChange  EventHandler
	onLogicChange   EventHandler

	// debounce state — protected by mu.
	mu     sync.Mutex
	timers map[string]*time.Timer
}

// New creates a new Watcher for the given root directory.
func New(ctx context.Context, root string, onSchema EventHandler, onLogic EventHandler) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	w := &Watcher{
		ctx:            ctx,
		root:           root,
		fsWatcher:      fsw,
		onSchemaChange: onSchema,
		onLogicChange:  onLogic,
		timers:         make(map[string]*time.Timer),
	}

	// Recursively add all directories under root.
	if err := w.addRecursive(root); err != nil {
		fsw.Close()
		return nil, err
	}

	return w, nil
}

// Start begins the event loop. It blocks until the context is cancelled.
func (w *Watcher) Start() error {
	defer w.close()

	fmt.Printf("👁  Watching: %s\n", w.root)
	fmt.Println("   Press Ctrl+C to stop.")
	fmt.Println()

	for {
		select {
		case <-w.ctx.Done():
			fmt.Println("\n🛑 Watcher stopped.")
			return nil

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return nil
			}
			w.handleEvent(event)

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "⚠  Watcher error: %v\n", err)
		}
	}
}

// handleEvent routes fsnotify events through the debouncer.
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Filter out noisy Chmod-only events.
	if event.Op == fsnotify.Chmod {
		return
	}

	// If a new directory is created, watch it recursively.
	if event.Has(fsnotify.Create) {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			if err := w.addRecursive(event.Name); err != nil {
				fmt.Fprintf(os.Stderr, "⚠  Failed to watch new directory %s: %v\n", event.Name, err)
			}
			return
		}
	}

	// Only process Write and Create events for files.
	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
		return
	}

	// Determine if this file is interesting.
	base := filepath.Base(event.Name)
	ext := strings.ToLower(filepath.Ext(event.Name))

	var handler EventHandler
	switch {
	case base == "schema.json":
		handler = w.onSchemaChange
	case ext == ".cs":
		handler = w.onLogicChange
	default:
		// Not a file we care about.
		return
	}

	// Debounce: reset the timer for this file path.
	w.mu.Lock()
	defer w.mu.Unlock()

	if t, exists := w.timers[event.Name]; exists {
		t.Stop()
	}

	path := event.Name
	h := handler
	w.timers[path] = time.AfterFunc(debounceInterval, func() {
		w.mu.Lock()
		delete(w.timers, path)
		w.mu.Unlock()

		// Check context before firing.
		if w.ctx.Err() != nil {
			return
		}

		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("\n[%s] 📝 Change detected: %s\n", timestamp, path)
		h(path)
	})
}

// addRecursive walks a directory tree and adds each directory to the watcher.
func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip hidden directories (e.g., .git).
			if strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}
			if err := w.fsWatcher.Add(path); err != nil {
				return fmt.Errorf("failed to watch %s: %w", path, err)
			}
		}
		return nil
	})
}

// close cleans up the watcher and all pending timers.
func (w *Watcher) close() {
	w.mu.Lock()
	for _, t := range w.timers {
		t.Stop()
	}
	w.timers = nil
	w.mu.Unlock()

	w.fsWatcher.Close()
}

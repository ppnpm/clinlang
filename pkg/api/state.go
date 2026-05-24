package api

import "sync"

// workspaceState is a thread-safe holder for the active workspace root.
// It is shared between the files handler and the workspace handler so
// that a successful PUT /api/v1/workspace immediately changes the root
// the next file operation will use.
//
// The struct lives on Server; handlers receive a pointer to the same
// instance so reads and writes go through one mutex.
type workspaceState struct {
	mu   sync.RWMutex
	path string
}

func (w *workspaceState) Get() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.path
}

func (w *workspaceState) Set(p string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.path = p
}

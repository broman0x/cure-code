package watcher

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher  *fsnotify.Watcher
	WorkDir  string
	OnChange func(path string, content string)
}

func NewWatcher(workDir string, onChange func(path string, content string)) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		watcher:  w,
		WorkDir:  workDir,
		OnChange: onChange,
	}, nil
}

func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					// Ignore specific dirs
					if isIgnored(event.Name) {
						continue
					}

					data, err := os.ReadFile(event.Name)
					if err == nil {
						w.OnChange(event.Name, string(data))
					}
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("[Watcher error] %v\n", err)
			}
		}
	}()

	// Add all subdirectories recursively
	filepath.WalkDir(w.WorkDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "target" || name == ".gemini" {
				return fs.SkipDir
			}
			w.watcher.Add(path)
		}
		return nil
	})
}

func isIgnored(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts {
		if p == ".git" || p == "node_modules" || p == "vendor" || p == "target" || p == ".gemini" {
			return true
		}
	}
	return false
}

func (w *Watcher) Close() error {
	return w.watcher.Close()
}

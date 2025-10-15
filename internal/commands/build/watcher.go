package build

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher *fsnotify.Watcher
}

func NewWatcher() *Watcher {
	return &Watcher{}
}

func (w *Watcher) Watch(buildFunc func()) error {
	var err error
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.watcher.Close()

	done := make(chan bool)
	var debounceTimer *time.Timer

	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(300*time.Millisecond, func() {
						fmt.Printf("\n🔄 File changed: %s\n", event.Name)
						buildFunc()
					})
				}
			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Watch error:", err)
			}
		}
	}()

	w.watcher.Add(".")
	if fileExists("src") {
		w.watcher.Add("./src")
	}
	if fileExists("hooks") {
		w.watcher.Add("./hooks")
	}
	if fileExists("components") {
		w.watcher.Add("./components")
	}

	fmt.Println("\n👀 Watching for changes...")
	<-done
	return nil
}

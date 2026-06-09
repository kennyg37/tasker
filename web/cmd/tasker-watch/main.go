// Command tasker-watch runs on the PC: it polls the local TODO.md and pushes
// changes to the API. The server cannot reach the PC, so capture is push-only.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/kennyg37/tasker/web/internal/config"
	"github.com/kennyg37/tasker/web/internal/watcher"
)

const pollInterval = 5 * time.Second

func main() {
	cfg := config.Load()
	if cfg.TodoFilePath == "" {
		log.Fatal("TODO_FILE_PATH is not set")
	}

	log.Printf("watching %s, syncing to %s every %s", cfg.TodoFilePath, cfg.APIURL, pollInterval)
	w := watcher.New(cfg.TodoFilePath, cfg.APIURL, cfg.SyncToken, http.DefaultClient)
	w.Run(context.Background(), pollInterval)
}

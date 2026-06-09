// Package watcher polls a TODO.md file and pushes its contents to the API
// whenever they change. It is a stdlib poller, not an event watcher: simple and
// dependency-free, which is all a single file needs.
package watcher

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Watcher struct {
	path     string
	apiURL   string
	token    string
	client   *http.Client
	lastHash [32]byte
	hasLast  bool
}

func New(path, apiURL, token string, client *http.Client) *Watcher {
	return &Watcher{path: path, apiURL: apiURL, token: token, client: client}
}

// Run polls until ctx is cancelled, syncing on every content change.
func (w *Watcher) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w.checkAndLog()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.checkAndLog()
		}
	}
}

func (w *Watcher) checkAndLog() {
	switch sent, err := w.checkOnce(); {
	case err != nil:
		log.Printf("watcher: %v", err)
	case sent:
		log.Printf("watcher: synced %s", w.path)
	}
}

// checkOnce reads the file and, if its content changed since the last sync,
// pushes it to the API. It reports whether a sync was sent.
func (w *Watcher) checkOnce() (bool, error) {
	data, err := os.ReadFile(w.path)
	if err != nil {
		return false, err
	}
	sum := sha256.Sum256(data)
	if w.hasLast && sum == w.lastHash {
		return false, nil
	}
	if err := w.push(data); err != nil {
		return false, err
	}
	w.lastHash = sum
	w.hasLast = true
	return true, nil
}

func (w *Watcher) push(data []byte) error {
	req, err := http.NewRequest(http.MethodPost, w.apiURL+"/api/sync", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	if w.token != "" {
		req.Header.Set("Authorization", "Bearer "+w.token)
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sync: status %d: %s", resp.StatusCode, body)
	}
	return nil
}

package watcher

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckOnceSyncsOnlyOnChange(t *testing.T) {
	var posts []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sync" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		posts = append(posts, string(b))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "TODO.md")
	if err := os.WriteFile(path, []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := New(path, srv.URL, "", srv.Client())

	// First read: pushes.
	if sent, err := w.checkOnce(); err != nil || !sent {
		t.Fatalf("first check: sent=%v err=%v, want sent", sent, err)
	}
	// Unchanged: no push.
	if sent, err := w.checkOnce(); err != nil || sent {
		t.Fatalf("unchanged check: sent=%v err=%v, want no push", sent, err)
	}
	// Changed: pushes again.
	if err := os.WriteFile(path, []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if sent, err := w.checkOnce(); err != nil || !sent {
		t.Fatalf("changed check: sent=%v err=%v, want push", sent, err)
	}

	want := []string{"v1", "v2"}
	if len(posts) != len(want) || posts[0] != want[0] || posts[1] != want[1] {
		t.Fatalf("posts = %v, want %v", posts, want)
	}
}

func TestPushSendsAuthHeaderWhenTokenSet(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "TODO.md")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := New(path, srv.URL, "topsecret", srv.Client())
	if _, err := w.checkOnce(); err != nil {
		t.Fatalf("checkOnce: %v", err)
	}
	if gotAuth != "Bearer topsecret" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer topsecret")
	}
}

func TestCheckOnceReturnsErrorOnMissingFile(t *testing.T) {
	w := New(filepath.Join(t.TempDir(), "nope.md"), "http://unused", "", http.DefaultClient)
	if _, err := w.checkOnce(); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPushReturnsErrorOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "TODO.md")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := New(path, srv.URL, "", srv.Client())
	if _, err := w.checkOnce(); err == nil {
		t.Fatal("expected error on non-200 response")
	}
	if w.hasLast {
		t.Error("hash should not be recorded when push fails (so it retries)")
	}
}

// Package store holds tasks. This is the Go equivalent of TodoManager.
//
// For now it is a thread-safe in-memory map so the app runs with zero external
// dependencies. Swap this implementation for SQLite later — the handlers only
// depend on the method set, not the backing storage.
package store

import (
	"sync"

	"github.com/kennyg37/tasker/web/internal/task"
)

type Store struct {
	mu     sync.RWMutex
	nextID int
	tasks  map[int]task.Task
}

func New() *Store {
	return &Store{
		nextID: 1,
		tasks:  make(map[int]task.Task),
	}
}

func (s *Store) Add(t task.Task) task.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	t.ID = s.nextID
	s.nextID++
	s.tasks[t.ID] = t
	return t
}

func (s *Store) All() []task.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]task.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	return out
}

func (s *Store) Get(id int) (task.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tasks[id]
	return t, ok
}

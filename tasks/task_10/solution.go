package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	R := &inMemoryTaskRepo{
		clock: clock,
		tasks: make(map[string]Task),
	}
	return R
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrInvalidTitle
	}
	T := Task{
		ID:        fmt.Sprintf("%d", r.seq),
		Title:     title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}
	r.seq++
	r.tasks[T.ID] = T
	return T, nil
}
func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	T, ok := r.tasks[id]
	return T, ok
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var tasks = make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		if t.Done == done {
			tasks = append(tasks, t)
		}
	}
	return tasks
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	T, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	if T.Done == done {
		return T, nil
	} else {
		T.Done = done
		T.UpdatedAt = r.clock.Now()
		r.tasks[id] = T
	}
	return T, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	H := &httpHandler{repo: repo}
	return H
}
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/tasks":
		switch r.Method {
		case http.MethodPost:
			h.handleCreate(w, r)
		case http.MethodGet:
			h.handleList(w, r)
		default:
			http.Error(w, "405", http.StatusMethodNotAllowed)
		}
	case strings.HasPrefix(r.URL.Path, "/tasks/"):
		id := strings.TrimPrefix(r.URL.Path, "/tasks/")
		if id == "" {
			http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		if strings.Contains(id, "/") {
			http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.handleGet(w, r, id)
		case http.MethodPatch:
			h.handlePatch(w, r, id)
		default:
			http.Error(w, "405", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
	}
}

func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	err := decodeStrictJSON(r.Body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err := h.repo.Create(req.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if !ok {
		http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, task)
}
func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	var done bool
	q := r.URL.Query()["done"]
	switch {
	case q == nil:
		done = false
	case len(q) > 1:
		http.Error(w, "list handler error > 1 value", http.StatusBadRequest)
		return
	case q[0] == "true":
		done = true
	case q[0] == "false":
		done = false
	default:
		http.Error(w, "list handler error'", http.StatusBadRequest)
		return
	}
	tasks := h.repo.List(done)
	slices.SortFunc(tasks, func(a, b Task) int {
		if b.UpdatedAt.Compare(a.UpdatedAt) != 0 {
			return b.UpdatedAt.Compare(a.UpdatedAt)
		}
		return strings.Compare(a.ID, b.ID)

	})
	writeJSON(w, http.StatusOK, tasks)
}
func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Done *bool `json:"done"`
	}
	err := decodeStrictJSON(r.Body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Done == nil {
		http.Error(w, "done field is required", http.StatusBadRequest)
		return
	}
	value := *req.Done
	task, err := h.repo.SetDone(id, value)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, task)
}
func decodeStrictJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	err := dec.Decode(v)
	if err != nil {
		return err
	}
	var extra any
	err = dec.Decode(&extra)
	if err != io.EOF {
		return errors.New("JSON error: trash after JSON object")
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

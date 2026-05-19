package api

import (
	"encoding/json"
	"net/http"

	"linkshelf/internal/store"
)

type Handlers struct {
	store store.Store
}

func NewHandlers(store store.Store) *Handlers {
	return &Handlers{store: store}
}

func (h *Handlers) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.store.ListTasks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handlers) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t store.Task
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.store.CreateTask(r.Context(), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/tasks/"):]
	task, err := h.store.GetTask(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(task)
}

func (h *Handlers) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t store.Task
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.store.UpdateTask(r.Context(), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/tasks/"):]
	err := h.store.DeleteTask(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

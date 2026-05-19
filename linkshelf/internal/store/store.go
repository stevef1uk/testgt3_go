package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID          string
	Title       string
	Description string
	Status      string
	CreatedAt   time.Time
}

func NewTask(title, description string) *Task {
	return &Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	s := &Store{db: db}
	s.initTable()
	return s
}

func (s *Store) initTable() {
	query := `
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		panic(fmt.Sprintf("failed to create table: %v", err))
	}
}

func (s *Store) CreateTask(ctx context.Context, task *Task) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tasks (id, title, description, status, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, task.ID, task.Title, task.Description, task.Status, task.CreatedAt)
	return err
}

func (s *Store) GetTask(ctx context.Context, id string) (*Task, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, title, description, status, created_at
		FROM tasks
		WHERE id = ?
	`, id)
	var t Task
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (s *Store) ListTasks(ctx context.Context) ([]*Task, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, description, status, created_at
		FROM tasks
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Store) UpdateTask(ctx context.Context, task *Task) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE tasks
		SET title = ?, description = ?, status = ?
		WHERE id = ?
	`, task.Title, task.Description, task.Status, task.ID)
	return err
}

func (s *Store) DeleteTask(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	return err
}

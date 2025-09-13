package domain

import "time"

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`

	IsEnd bool `json:"is_end"`

	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"updated_at"`
	LimitedAt time.Time `json:"limited_at"`
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TaskTag struct {
	TaskID string `json:"task_id"`
	TagID  string `json:"tag_id"`
}

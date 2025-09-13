package domain

import "time"

type CreateTaskParam struct {
	ID          string    `json:"id"`
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	LimitedAt   time.Time `json:"limited_at"`
	IsEnd       bool      `json:"is_end"`

	TagIDs []string `json:"tag_ids"`
}

type GetTaskParam struct {
	ID string `json:"id"`
}

type ListTaskParam struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type UpdateTaskParam struct {
	ID          string    `json:"id" validate:"required"`
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	LimitedAt   time.Time `json:"limited_at"`
	IsEnd       bool      `json:"is_end"`
}

type DeleteTaskParam struct {
	ID string `json:"id"`
}

type CreateTagParam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GetTagParam struct {
	ID string `json:"id"`
}

type ListTagParam struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type DeleteTagParam struct {
	ID string `json:"id"`
}

type CreateTaskTagParam struct {
	TaskID string `json:"task_id"`
	TagID  string `json:"tag_id"`
}

type GetTaskTagParam struct {
	TaskID string `json:"task_id"`
}

type DeleteTaskTagParam struct {
	TaskID string `json:"task_id"`
}

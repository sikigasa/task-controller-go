package infra

import (
	"context"
	"database/sql"

	"github.com/sikigasa/task-controller/internal/domain"
)

type taskTagRepo struct {
	db *sql.DB
}

type TaskTagRepo interface {
	CreateTaskTag(ctx context.Context, tx *sql.Tx, arg domain.CreateTaskTagParam) error
	GetTaskTagIDs(ctx context.Context, arg domain.GetTaskTagParam) ([]domain.TaskTag, error)
	DeleteTaskTags(ctx context.Context, tx *sql.Tx, arg domain.DeleteTaskTagParam) error
}

func NewTaskTagRepo(db *sql.DB) TaskTagRepo {
	return &taskTagRepo{db: db}
}

func (t *taskTagRepo) CreateTaskTag(ctx context.Context, tx *sql.Tx, arg domain.CreateTaskTagParam) error {
	const query = `INSERT INTO task_tag (task_id, tag_id) VALUES ($1,$2)`

	_, err := tx.ExecContext(ctx, query, arg.TaskID, arg.TagID)

	return err
}

func (t *taskTagRepo) GetTaskTagIDs(ctx context.Context, arg domain.GetTaskTagParam) ([]domain.TaskTag, error) {
	const query = `SELECT * FROM task_tag WHERE task_id = $1`

	rows, err := t.db.QueryContext(ctx, query, arg.TaskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taskTags []domain.TaskTag
	for rows.Next() {
		var taskTag domain.TaskTag
		if err := rows.Scan(&taskTag.TaskID, &taskTag.TagID); err != nil {
			return nil, err
		}
		taskTags = append(taskTags, taskTag)
	}
	return taskTags, nil
}

func (t *taskTagRepo) DeleteTaskTags(ctx context.Context, tx *sql.Tx, arg domain.DeleteTaskTagParam) error {
	const query = `DELETE FROM task_tag WHERE task_id = $1`
	_, err := tx.ExecContext(ctx, query, arg.TaskID)

	return err
}

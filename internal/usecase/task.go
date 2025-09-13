package usecase

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/sikigasa/task-controller/internal/domain"
	"github.com/sikigasa/task-controller/internal/infra"
	postgres "github.com/sikigasa/task-controller/internal/infra/driver"
	task "github.com/sikigasa/task-controller/proto/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type taskService struct {
	task.UnimplementedTaskServiceServer
	taskRepo    infra.TaskRepo
	tagRepo     infra.TagRepo
	taskTagRepo infra.TaskTagRepo
	tx          postgres.Transaction
}

func NewTaskService(taskRepo infra.TaskRepo, tagRepo infra.TagRepo, taskTagRepo infra.TaskTagRepo, tx postgres.Transaction) task.TaskServiceServer {
	return &taskService{
		taskRepo:    taskRepo,
		tagRepo:     tagRepo,
		taskTagRepo: taskTagRepo,
		tx:          tx,
	}
}

func (t *taskService) CreateTask(ctx context.Context, req *task.CreateTaskRequest) (*task.CreateTaskResponse, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	err = t.tx.WithTransaction(ctx, func(tx *sql.Tx) error {
		param := domain.CreateTaskParam{
			ID:          uuid.String(),
			Title:       req.Title,
			Description: req.Description,
			LimitedAt:   req.LimitedAt.AsTime(),
			IsEnd:       false,
		}

		if err := t.taskRepo.CreateTask(ctx, tx, param); err != nil {
			return err
		}

		if len(req.TagIds) == 0 || req.TagIds[0] == "" {
			return nil
		}
		for _, tagID := range req.TagIds {
			taskTagParam := domain.CreateTaskTagParam{
				TaskID: param.ID,
				TagID:  tagID,
			}
			if err := t.taskTagRepo.CreateTaskTag(ctx, tx, taskTagParam); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &task.CreateTaskResponse{
		Id: uuid.String(),
	}, nil
}

func (t *taskService) GetTask(ctx context.Context, req *task.GetTaskRequest) (*task.GetTaskResponse, error) {
	param := domain.GetTaskParam{
		ID: req.Id,
	}

	taskDetail, err := t.taskRepo.GetTask(ctx, param)
	if err != nil {
		return nil, err
	}

	taskTagIDs, err := t.taskTagRepo.GetTaskTagIDs(ctx, domain.GetTaskTagParam{TaskID: taskDetail.ID})
	if err != nil {
		return nil, err
	}
	var protoTags []*task.Tag
	for _, tagID := range taskTagIDs {
		tag, err := t.tagRepo.GetTag(ctx, domain.GetTagParam{ID: tagID.TagID})
		if err != nil {
			return nil, err
		}
		protoTags = append(protoTags, &task.Tag{
			Id:   tag.ID,
			Name: tag.Name,
		})
	}

	return &task.GetTaskResponse{
		Task: &task.Task{
			Id:          taskDetail.ID,
			Title:       taskDetail.Title,
			Description: taskDetail.Description,
			CreatedAt:   timestamppb.New(taskDetail.CreatedAt),
			UpdatedAt:   timestamppb.New(taskDetail.UpdateAt),
			LimitedAt:   timestamppb.New(taskDetail.LimitedAt),
			IsEnd:       taskDetail.IsEnd,
			Tags:        protoTags,
		},
	}, nil
}

func (t *taskService) ListTask(ctx context.Context, req *task.ListTaskRequest) (*task.ListTaskResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}
	param := domain.ListTaskParam{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	tasks, err := t.taskRepo.ListTask(ctx, param)
	if err != nil {
		return nil, err
	}

	var taskList []*task.Task
	for _, taskDetail := range tasks {
		taskTagIDs, err := t.taskTagRepo.GetTaskTagIDs(ctx, domain.GetTaskTagParam{TaskID: taskDetail.ID})
		if err != nil {
			return nil, err
		}
		var protoTags []*task.Tag
		for _, tagID := range taskTagIDs {
			tag, err := t.tagRepo.GetTag(ctx, domain.GetTagParam{ID: tagID.TagID})
			if err != nil {
				return nil, err
			}
			protoTags = append(protoTags, &task.Tag{
				Id:   tag.ID,
				Name: tag.Name,
			})
		}

		taskList = append(taskList, &task.Task{
			Id:          taskDetail.ID,
			Title:       taskDetail.Title,
			Description: taskDetail.Description,
			CreatedAt:   timestamppb.New(taskDetail.CreatedAt),
			UpdatedAt:   timestamppb.New(taskDetail.UpdateAt),
			LimitedAt:   timestamppb.New(taskDetail.LimitedAt),
			IsEnd:       taskDetail.IsEnd,
			Tags:        protoTags,
		})
	}

	return &task.ListTaskResponse{
		Tasks: taskList,
	}, nil
}

func (t *taskService) UpdateTask(ctx context.Context, req *task.UpdateTaskRequest) (*task.UpdateTaskResponse, error) {
	if req.TagIds == nil {
		req.TagIds = []string{}
	}
	err := t.tx.WithTransaction(ctx, func(tx *sql.Tx) error {
		param := domain.UpdateTaskParam{
			ID:          req.Id,
			Title:       req.Title,
			Description: req.Description,
			LimitedAt:   req.LimitedAt.AsTime(),
			IsEnd:       req.IsEnd,
		}
		if err := t.taskRepo.UpdateTask(ctx, tx, param); err != nil {
			return err
		}
		if err := t.taskTagRepo.DeleteTaskTags(ctx, tx, domain.DeleteTaskTagParam{TaskID: req.Id}); err != nil {
			return err
		}

		if len(req.TagIds) == 0 || req.TagIds[0] == "" {
			return nil
		}
		for _, tagID := range req.TagIds {
			taskTagParam := domain.CreateTaskTagParam{
				TaskID: param.ID,
				TagID:  tagID,
			}
			if err := t.taskTagRepo.CreateTaskTag(ctx, tx, taskTagParam); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &task.UpdateTaskResponse{
		Success: true,
	}, nil
}

func (t *taskService) DeleteTask(ctx context.Context, req *task.DeleteTaskRequest) (*task.DeleteTaskResponse, error) {
	err := t.tx.WithTransaction(ctx, func(tx *sql.Tx) error {
		if err := t.taskTagRepo.DeleteTaskTags(ctx, tx, domain.DeleteTaskTagParam{TaskID: req.Id}); err != nil {
			return err
		}

		param := domain.DeleteTaskParam{
			ID: req.Id,
		}
		if err := t.taskRepo.DeleteTask(ctx, tx, param); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &task.DeleteTaskResponse{
		Success: true,
	}, nil
}

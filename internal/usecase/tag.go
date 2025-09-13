package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/sikigasa/task-controller/internal/domain"
	"github.com/sikigasa/task-controller/internal/infra"
	tag "github.com/sikigasa/task-controller/proto/v1"
)

type TagService struct {
	tag.UnimplementedTagServiceServer
	tagRepo infra.TagRepo
}

func NewTagService(tagRepo infra.TagRepo) tag.TagServiceServer {
	return &TagService{
		tagRepo: tagRepo,
	}
}

func (t *TagService) CreateTag(ctx context.Context, req *tag.CreateTagRequest) (*tag.CreateTagResponse, error) {
	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	param := domain.CreateTagParam{
		ID:   uuid.String(),
		Name: req.Name,
	}

	if err := t.tagRepo.CreateTag(ctx, param); err != nil {
		return nil, err
	}

	return &tag.CreateTagResponse{
		Id: param.ID,
	}, nil
}

func (t *TagService) ListTag(ctx context.Context, req *tag.ListTagRequest) (*tag.ListTagResponse, error) {
	param := domain.ListTagParam{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	tags, err := t.tagRepo.ListTag(ctx, param)
	if err != nil {
		return nil, err
	}

	var tagList []*tag.Tag
	for _, t := range tags {
		tagList = append(tagList, &tag.Tag{
			Id:   t.ID,
			Name: t.Name,
		})
	}

	return &tag.ListTagResponse{
		Tags: tagList,
	}, nil
}

func (t *TagService) DeleteTag(ctx context.Context, req *tag.DeleteTagRequest) (*tag.DeleteTagResponse, error) {
	param := domain.DeleteTagParam{
		ID: req.Id,
	}

	if err := t.tagRepo.DeleteTag(ctx, param); err != nil {
		return nil, err
	}

	return &tag.DeleteTagResponse{
		Success: true,
	}, nil
}

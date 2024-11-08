package model

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/usesend0/send0/internal/uid"
)

type TagRepository interface {
	Save(ctx context.Context, tag *Tag) error
	SaveMany(ctx context.Context, tags []*Tag) error
	FindByWorkspaceId(ctx context.Context, workspaceId uid.UID) ([]*Tag, error)
}

type Tag struct {
	Base
	Name        string  `json:"name"`
	WorkspaceId uid.UID `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type tagRepository struct {
	*baseRepository
}

func NewTagRepository(baseRepository *baseRepository) TagRepository {
	return &tagRepository{baseRepository}
}

func (r *tagRepository) Save(ctx context.Context, tag *Tag) error {
	stmt, args, err := r.DB.Builder().Insert(string(TableNameTag)).Columns(
		"id",
		"name",
		"workspace_id",
	).Values(
		tag.Id,
		tag.Name,
		tag.WorkspaceId,
	).ToSql()
	if err != nil {
		return err
	}
	_, err = r.DB.Connection().Exec(ctx, stmt, args...)

	return err
}

func (r *tagRepository) SaveMany(ctx context.Context, tags []*Tag) error {
	_, err := r.DB.Connection().CopyFrom(
		ctx,
		pgx.Identifier{string(TableNameTag)},
		[]string{"id", "name", "workspace_id"},
		pgx.CopyFromSlice(len(tags), func(i int) ([]interface{}, error) {
			return []interface{}{
				tags[i].Id,
				tags[i].Name,
				tags[i].WorkspaceId,
			}, nil
		}),
	)

	return err
}

func (r *tagRepository) FindByWorkspaceId(ctx context.Context, workspaceId uid.UID) ([]*Tag, error) {
	var tags []*Tag
	stmt := `SELECT * FROM tags WHERE workspace_id = $1`
	rows, err := r.DB.Connection().Query(ctx, stmt, workspaceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}

	return tags, nil
}

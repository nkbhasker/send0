package model

import (
	"context"

	"github.com/usesend0/send0/internal/uid"
)

type SegmentRepository interface {
	Create(ctx context.Context, segment *Segment) error
	FindByID(ctx context.Context, id uid.UID) (*Segment, error)
}

type Segment struct {
	Base
	Name           string     `json:"name"`
	Description    *string    `json:"description omitempty" db:"description"`
	IsDefault      bool       `json:"isDefault" db:"is_default" gorm:"not null;default:false"`
	IsPrivate      bool       `json:"isPrivate" db:"is_private" gorm:"not null;default:false"`
	TotalCount     int        `json:"totalCount" db:"total_count" gorm:"not null;default:0"`
	Tags           JSONBArray `json:"tags" db:"tags" gorm:"type:jsonb;not null;default '[]'"`
	OrganizationId uid.UID    `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID    `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type SegmentContact struct {
	Base
	SegmentId      uid.UID `json:"segmentId" db:"segment_id" gorm:"not null"`
	ContactId      uid.UID `json:"contactId" db:"contact_id" gorm:"not null"`
	Subscribed     bool    `json:"subscribed" db:"subscribed" gorm:"not null;default:false"`
	OrganizationId uid.UID `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type segmentRepository struct {
	*baseRepository
}

func NewSegmentRepository(baseRepository *baseRepository) SegmentRepository {
	return &segmentRepository{
		baseRepository,
	}
}

func (r *segmentRepository) Create(ctx context.Context, segment *Segment) error {
	stmt := `INSERT INTO segments (
		id,
		name,
		description,
		tags,
		is_private,
		total_count,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		segment.Id,
		segment.Name,
		segment.Description,
		segment.Tags,
		segment.IsPrivate,
		segment.TotalCount,
		segment.WorkspaceId,
	)

	return err
}

func (r *segmentRepository) FindByID(ctx context.Context, id uid.UID) (*Segment, error) {
	var segment Segment
	stmt := `SELECT * FROM segments WHERE id = $1`
	err := r.DB.Connection().QueryRow(ctx, stmt, id).Scan(&segment)
	if err != nil {
		return nil, err
	}

	return &segment, nil
}

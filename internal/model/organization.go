package model

import (
	"context"

	"github.com/usesend0/send0/internal/uid"
)

type OrganizationRepository interface {
	Create(ctx context.Context, organization *Organization) error
	FindById(ctx context.Context, id uid.UID) (*Organization, error)
	FindAll(ctx context.Context) ([]*Organization, int, error)
}
type Organization struct {
	Base
	Name                string     `json:"name" db:"name" gorm:"not null"`
	IsDefault           bool       `json:"isDefault" db:"is_default" gorm:"not null;default:false"`
	WorkspaceId         uid.UID    `json:"workspaceId" db:"workspace_id" gorm:"not null"`
	CCAddresses         JSONBArray `json:"ccAddresses" db:"cc_addresses" gorm:"type:jsonb;not null;default:'[]'"`
	Subdomain           string     `json:"subdomain" db:"subdomain" gorm:"not null"`
	AutoOptIn           bool       `json:"autoOptInEmail" db:"auto_opt_in_email" gorm:"not null;default:false"`
	AutoOptInTemplateId uid.UID    `json:"autoOptInTemplateId" db:"auto_opt_in_template_id"`
}

type organizationRepository struct {
	*baseRepository
}

func NewOrganizationRepository(baseRepository *baseRepository) OrganizationRepository {
	return &organizationRepository{baseRepository}
}

func (r *organizationRepository) Create(ctx context.Context, organization *Organization) error {
	stmt := `INSERT INTO organizations (
		id,
		name,
		is_default,
		cc_addresses,
		subdomain,
		auto_opt_in_template_id,
		workspace_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.DB.Connection().Exec(
		ctx,
		stmt,
		r.UID(organization.Id),
		organization.Name,
		organization.IsDefault,
		organization.CCAddresses,
		organization.Subdomain,
		organization.AutoOptInTemplateId,
		organization.WorkspaceId,
	)

	return err
}

func (r *organizationRepository) FindById(ctx context.Context, id uid.UID) (*Organization, error) {
	stmt := `SELECT
		id,
		name,
		is_default,
		workspace_id,
		cc_addresses
	FROM organizations WHERE id = $1`

	var organization Organization
	err := r.DB.Connection().QueryRow(ctx, stmt, id).Scan(&organization)

	return &organization, err
}

func (r *organizationRepository) FindAll(ctx context.Context) ([]*Organization, int, error) {
	var count int
	err := r.DB.Connection().QueryRow(ctx, "SELECT COUNT(*) FROM organizations").Scan()
	if err != nil {
		return nil, count, err
	}
	stmt := `SELECT
		id,
		name,
		is_default,
		workspace_id,
		cc_addresses
	FROM organizations`
	rows, err := r.DB.Connection().Query(ctx, stmt)
	if err != nil {
		return nil, count, err
	}

	var organizations []*Organization
	for rows.Next() {
		var organization Organization
		err := rows.Scan(&organization)
		if err != nil {
			return nil, 0, err
		}

		organizations = append(organizations, &organization)
	}

	return organizations, count, nil
}

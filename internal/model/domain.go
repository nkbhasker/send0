package model

import (
	"context"
	"fmt"

	"github.com/usesend0/send0/internal/constant"
	"github.com/usesend0/send0/internal/uid"
)

var DomainStatusTypeCreateQuery = fmt.Sprintf(
	`CREATE TYPE %s AS ENUM ('%s','%s','%s');`,
	DBTypeDomainStatus,
	constant.DomainStatusPending,
	constant.DomainStatusActive,
	constant.DomainStatusInactive,
)

type DomainRepository interface {
	Save(ctx context.Context, domain *Domain) error
	FindAll(ctx context.Context, options DomainFindOptions) ([]*Domain, error)
	FindById(ctx context.Context, id uid.UID) (*Domain, error)
	FindByDomainName(ctx context.Context, workspaceId, organizationId uid.UID, domainName string) (*Domain, error)
	Delete(ctx context.Context, id uid.UID) error
}

type Domain struct {
	Base
	Name           string                     `json:"name" db:"name" gorm:"not null"`
	Region         constant.AwsRegion         `json:"region" db:"region" gorm:"not null"`
	Status         constant.DomainStatus      `json:"status" db:"status" gorm:"not null;type:domain_status;default:'PENDING'"`
	DKIMRecords    constant.JSONDomainRecords `json:"dkimRecords" db:"dkim_records" gorm:"type:jsonb;not null;default '[]'"`
	SPFRecords     constant.JSONDomainRecords `json:"spfRecords" db:"spf_records" gorm:"type:jsonb;not null;default '[]'"`
	DMARCRecords   constant.JSONDomainRecords `json:"dmarcRecords" db:"dmarc_records" gorm:"type:jsonb;not null;default '[]'"`
	PrivateKey     JSONPrivateKey             `json:"-" db:"private_key" gorm:"not null"`
	OrganizationId uid.UID                    `json:"organizationId" db:"organization_id" gorm:"not null"`
	WorkspaceId    uid.UID                    `json:"workspaceId" db:"workspace_id" gorm:"not null"`
}

type DomainFindOptions struct {
	Id             uid.UID
	Name           string
	Region         constant.AwsRegion
	OrganizationId uid.UID
	WorkspaceId    uid.UID
}

type domainRepository struct {
	*baseRepository
}

func NewDomainRepository(baseRepository *baseRepository) DomainRepository {
	return &domainRepository{
		baseRepository,
	}
}

func (r *domainRepository) Save(ctx context.Context, domain *Domain) error {
	stmt, args, err := r.DB.Builder().Insert(string(TableNameDomain)).Columns(
		"id",
		"name",
		"region",
		"status",
		"dkim_records",
		"spf_records",
		"dmarc_records",
		"private_key",
		"organization_id",
		"workspace_id",
	).Values(
		domain.Id,
		domain.Name,
		domain.Region,
		domain.Status,
		domain.DKIMRecords,
		domain.SPFRecords,
		domain.DMARCRecords,
		domain.PrivateKey,
		domain.OrganizationId,
		domain.WorkspaceId,
	).ToSql()
	if err != nil {
		return err
	}

	_, err = r.DB.Connection().Exec(ctx, stmt, args...)
	return err
}

func (r *domainRepository) FindAll(ctx context.Context, options DomainFindOptions) ([]*Domain, error) {
	var domains []*Domain
	stmt, args, err := r.DB.Builder().Select(
		"id",
		"updated_at",
		"name",
		"region",
		"status",
		"organization_id",
		"workspace_id",
	).From(string(TableNameDomain)).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.Connection().Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var domain Domain
		err = rows.Scan(
			&domain.Id,
			&domain.UpdatedAt,
			&domain.Name,
			&domain.Region,
			&domain.Status,
			&domain.OrganizationId,
			&domain.WorkspaceId,
		)
		if err != nil {
			return nil, err
		}
		domains = append(domains, &domain)
	}

	return domains, nil
}

func (r *domainRepository) FindById(ctx context.Context, id uid.UID) (*Domain, error) {
	var domain Domain
	stmt, args, err := r.DB.Builder().Select("*").From(string(TableNameDomain)).Where("id = ?", id).ToSql()
	if err != nil {
		return nil, err
	}
	err = r.DB.Connection().QueryRow(ctx, stmt, args...).Scan(
		&domain.Id,
		&domain.UpdatedAt,
		&domain.Name,
		&domain.Region,
		&domain.Status,
		&domain.DKIMRecords,
		&domain.SPFRecords,
		&domain.DMARCRecords,
		&domain.PrivateKey,
		&domain.OrganizationId,
		&domain.WorkspaceId,
	)

	return &domain, err
}

func (r *domainRepository) FindByDomainName(
	ctx context.Context,
	workspaceId,
	organizationId uid.UID,
	domainName string,
) (*Domain, error) {
	var domain Domain
	stmt, args, err := r.DB.Builder().Select("*").From(string(TableNameDomain)).
		Where("name = ?", domainName).
		Where("workspace_id = ?", workspaceId).
		Where("organization_id = ?", organizationId).
		ToSql()
	if err != nil {
		return nil, err
	}
	err = r.DB.Connection().QueryRow(ctx, stmt, args...).Scan(
		&domain.Id,
		&domain.UpdatedAt,
		&domain.Name,
		&domain.Region,
		&domain.Status,
		&domain.DKIMRecords,
		&domain.SPFRecords,
		&domain.DMARCRecords,
		&domain.PrivateKey,
		&domain.OrganizationId,
		&domain.WorkspaceId,
	)

	return &domain, err
}

func (r *domainRepository) Delete(ctx context.Context, id uid.UID) error {
	stmt, args, err := r.DB.Builder().Delete(string(TableNameDomain)).Where("id = ?", id).ToSql()
	if err != nil {
		return err
	}

	_, err = r.DB.Connection().Exec(ctx, stmt, args...)
	return err
}

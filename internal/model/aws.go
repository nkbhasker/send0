package model

import (
	"context"

	"github.com/usesend0/send0/internal/constant"
)

type SNSTopicRepository interface {
	Save(ctx context.Context, topic *SNSTopic) error
	FindAll(ctx context.Context) ([]*SNSTopic, error)
	UpdateStatus(ctx context.Context, region constant.AwsRegion, status constant.AwsSNSTopicStatus) error
}

type SNSTopic struct {
	Base
	Region constant.AwsRegion         `json:"region" db:"region" gorm:"not null"`
	Arn    string                     `json:"arn"`
	Status constant.AwsSNSTopicStatus `json:"status"`
}

type snsTopicRepository struct {
	*baseRepository
}

func NewSNSTopicRepository(baseRepository *baseRepository) SNSTopicRepository {
	return &snsTopicRepository{
		baseRepository,
	}
}

func (r *snsTopicRepository) Save(ctx context.Context, topic *SNSTopic) error {
	stmt, args, err := r.DB.Builder().Insert(string(TableNameSNSTopic)).Columns(
		"id",
		"region",
		"arn",
		"status",
	).Values(
		r.UID(topic.Id),
		topic.Region,
		topic.Arn,
		topic.Status,
	).ToSql()
	if err != nil {
		return err
	}
	_, err = r.DB.Connection().Exec(ctx, stmt, args...)
	return err
}

func (r *snsTopicRepository) FindAll(ctx context.Context) ([]*SNSTopic, error) {
	var topics []*SNSTopic
	stmt, args, err := r.DB.Builder().Select(
		"id",
		"region",
		"arn",
		"status",
	).From(string(TableNameSNSTopic)).ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.Connection().Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var topic SNSTopic
		if err := rows.Scan(
			&topic.Id,
			&topic.Region,
			&topic.Arn,
			&topic.Status,
		); err != nil {
			return nil, err
		}
		topics = append(topics, &topic)
	}

	return topics, nil
}

func (r *snsTopicRepository) UpdateStatus(ctx context.Context, region constant.AwsRegion, status constant.AwsSNSTopicStatus) error {
	stmt, args, err := r.DB.Builder().Update(string(TableNameSNSTopic)).Set("status", status).Where(
		"region = ?", region,
	).ToSql()
	if err != nil {
		return err
	}
	_, err = r.DB.Connection().Exec(ctx, stmt, args...)
	return err
}

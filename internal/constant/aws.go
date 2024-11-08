package constant

import (
	"database/sql"
	"database/sql/driver"
)

const (
	AwsRegionNorthVirginia AwsRegion = "us-east-1"
	AwsRegionIreland       AwsRegion = "eu-west-1"
	AwsRegionSaopaulo      AwsRegion = "sa-east-1"
	AwsRegionTokyo         AwsRegion = "ap-northeast-1"
)

const (
	AwsSNSTopicStatusActive   AwsSNSTopicStatus = "ACTIVE"
	AwsSNSTopicStatusInactive AwsSNSTopicStatus = "INACTIVE"
	AwsSNSTopicStatusPending  AwsSNSTopicStatus = "PENDING"
)

const AwsSESEventTopicSuccessMessage string = "Successfully validated SNS topic for Amazon SES event publishing."

var SupportedSESRegions = []AwsRegion{
	AwsRegionNorthVirginia,
	AwsRegionIreland,
	AwsRegionSaopaulo,
	AwsRegionTokyo,
}

var _ sql.Scanner = (*AwsRegion)(nil)
var _ driver.Valuer = (*AwsRegion)(nil)

type AwsRegion string
type AwsSNSTopicStatus string
type AwsSESEventType string

func (r *AwsRegion) Scan(value interface{}) error {
	*r = AwsRegion(value.(string))
	return nil
}

func (r AwsRegion) Value() (driver.Value, error) {
	return string(r), nil
}

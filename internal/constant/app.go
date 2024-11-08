package constant

import "time"

const (
	EnvDevelopment Env = "DEVELOPMENT"
	EnvStaging     Env = "STAGING"
	EnvProduction  Env = "PRODUCTION"
)

const DialectPostgres = "postgres"

const AppName = "send0"
const SNSEventPath = "/sns/events"
const (
	HeaderAuthorization = "authorization"
	HeaderWorkspaceId   = "x-workspace-id"
	HeaderXFowardedFor  = "x-forwarded-for"
)

var RetrySchedule = []time.Duration{
	0 * time.Second,
	5 * time.Second,
	1 * time.Minute,
	5 * time.Minute,
	10 * time.Minute,
	30 * time.Minute,
	1 * time.Hour,
	5 * time.Hour,
	10 * time.Hour,
	24 * time.Hour,
}

type Env string

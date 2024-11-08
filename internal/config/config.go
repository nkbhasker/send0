package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/usesend0/send0/internal/constant"
)

type Config struct {
	Host           string       `default:"http://localhost:8000"`
	Port           string       `default:"8000"`
	Postgres       Postgres     `required:"true"`
	Redis          Redis        `required:"true"`
	Authn          Authn        `required:"true"`
	SES            SES          `required:"true"`
	S3             S3           `required:"true"`
	SNS            SNS          `required:"true"`
	Env            constant.Env `default:"DEVELOPMENT"`
	JWT            JWT          `required:"true"`
	AdminEmail     string       `required:"true" default:"admin@send0.com"`
	WorkspaceId    int          `default:"123456789"`
	OrganizationId int          `default:"123456789"`
}

type Postgres struct {
	URL          string `required:"true" default:"postgres://send0:send0@localhost:5432/send0?sslmode=disable"`
	PoolSize     int    `default:"10"`
	IdlePoolSize int    `default:"5"`
}

type Redis struct {
	URL string `required:"true" default:"redis://localhost:6379"`
}

type Authn struct {
	OtpExpiryInMinutes         int `default:"5"`
	OtpGenerateRateLimit       int `default:"3"`
	OtpGenerateRateLimitWindow int `default:"7200"`
	OtpVerifyRateLimit         int `default:"5"`
	OtpVerifyRateLimitWindow   int `default:"86400"`
}

type SES struct {
	AccessKeyId     string `required:"true"`
	SecretAccessKey string `required:"true"`
}

type SNS struct {
	AccessKeyId     string `required:"true"`
	SecretAccessKey string `required:"true"`
	EndPoint        string `required:"false"`
}

type S3 struct {
	Region          string `default:"ap-south-1"`
	AccessKeyId     string `required:"true"`
	SecretAccessKey string `required:"true"`
}

type JWT struct {
	PrivateKey        string `required:"true"`
	AccessTokenExpiry int    `default:"1440"`
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	cnf := new(Config)
	err = envconfig.Process(constant.AppName, cnf)
	if err != nil {
		return nil, err
	}
	cnf.SNS.EndPoint = cnf.Host + constant.SNSEventPath

	return cnf, nil
}

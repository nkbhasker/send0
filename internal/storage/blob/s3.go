package blob

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/usesend0/send0/internal/config"
)

var _ BlobStorage = (*AWSS3)(nil)

type AWSS3 struct {
	svc *s3.S3
}

func NewAWSS3(config *config.Config) (*AWSS3, error) {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.S3.Region),
		Credentials: credentials.NewStaticCredentials(config.S3.AccessKeyId, config.S3.SecretAccessKey, ""),
	})
	if err != nil {
		return nil, err
	}
	svc := s3.New(session)
	return &AWSS3{
		svc: svc,
	}, nil
}

func (a *AWSS3) Get(id int) error {
	object, err := a.svc.GetObject(&s3.GetObjectInput{
		Key: aws.String(strconv.Itoa(id)),
	})
	if err != nil {
		return err
	}
	_ = object

	return nil
}

func (a *AWSS3) GetSignedURL(id int) (*string, error) {
	req, _ := a.svc.GetObjectRequest(&s3.GetObjectInput{
		Key: aws.String(strconv.Itoa(id)),
	})
	url, err := req.Presign(15 * time.Minute)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

package blob

import "github.com/usesend0/send0/internal/config"

type BlobStorage interface {
	Get(id int) error
	GetSignedURL(id int) (*string, error)
}

func NewBlob(config config.Config) (BlobStorage, error) {
	return NewAWSS3(&config)
}

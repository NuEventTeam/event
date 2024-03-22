package assets

import (
	"context"
	"errors"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/http"
	"time"
)

type Assets struct {
	client          *s3.Client        `json:"client,omitempty"`
	presignedClient *s3.PresignClient `json:"presignedClient,omitempty"`
	bucket          string            `json:"bucket,omitempty"`
}

func NewS3Storage(cfg *config.CDN) *Assets {

	client := s3.New(s3.Options{
		Region:      cfg.Region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.KeyID, cfg.SecretAccessKey, "")),
	})

	presignerClient := s3.NewPresignClient(client)

	return &Assets{
		client:          client,
		presignedClient: presignerClient,
		bucket:          cfg.BucketName,
	}
}

func (s *Assets) GetObjectURL(objectKey string) (*v4.PresignedHTTPRequest, error) {

	request, err := s.presignedClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(time.Hour*12))

	if err != nil {
		return nil, err
	}

	return request, err
}

func (s *Assets) KeyExists(ctx context.Context, objectKey string) bool {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return false
		}
		return false
	}
	return true
}

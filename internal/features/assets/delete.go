package assets

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
)

func (s *Assets) DeleteFile(ctx context.Context, path ...string) error {

	err := s.delete(ctx, path...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Assets) delete(ctx context.Context, keys ...string) error {
	for _, k := range keys {
		k := k
		go func() {
			_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    aws.String(k),
			})
			if err != nil {
				log.Println("error while deleting images")
			}
		}()

	}
	return nil
}

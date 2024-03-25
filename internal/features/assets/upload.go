package assets

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
)

func (s *Assets) Upload(ctx context.Context, images ...*Image) {

	for _, img := range images {
		img := img
		go func() {
			_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:             aws.String(s.bucket),
				Key:                aws.String(img.Filename),
				Body:               img.file,
				ContentDisposition: aws.String(fmt.Sprintf("inline; filename=%s", img.Filename)),
				ContentType:        aws.String(GetContentType[img.ext]),
			})
			if err != nil {
				log.Println("error while uploading to s3")
			}
		}()

	}
}

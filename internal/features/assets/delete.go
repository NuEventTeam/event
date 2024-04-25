package assets

import (
	"context"
	"github.com/NuEventTeam/events/pkg"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

func (s *Assets) DeleteFile(ctx context.Context, path ...string) error {

	err := s.delete(ctx, path...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Assets) delete(ctx context.Context, keys ...string) error {
	wg := &sync.WaitGroup{}
	wg.Add(len(keys))

	for _, k := range keys {
		k := k
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			k = strings.TrimPrefix(k, pkg.CDNBaseUrl)
			err := os.Remove(path.Join("static", k))
			if err != nil {
				log.Println("while deleting message", err)
			}
			//_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			//	Bucket: aws.String(s.bucket),
			//	Key:    aws.String(k),
			//})
			//if err != nil {
			//	log.Println("error while deleting images")
			//}
		}(wg)

	}
	return nil
}

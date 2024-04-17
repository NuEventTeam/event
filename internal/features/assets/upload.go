package assets

import (
	"bufio"
	"context"
	"log"
	"os"
	"path"
)

func (s *Assets) Upload(ctx context.Context, images ...Image) {
	log.Println(images)
	for _, img := range images {
		img := img
		go func() {
			if img.Filename == nil {
				return
			}
			//_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
			//	Bucket:             aws.String(s.bucket),
			//	Key:                img.Filename,
			//	Body:               img.file,
			//	ContentDisposition: aws.String(fmt.Sprintf("inline; filename=%s", img.Filename)),
			//	ContentType:        aws.String(GetContentType[img.ext]),
			//})

			absPath := path.Join("./static", *img.Filename)
			dir, _ := path.Split(absPath)
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				log.Println(err)
				return
			}
			file, err := os.Create(absPath)
			if err != nil {
				log.Println(err)
				return
			}
			//io.W
			writer := bufio.NewWriter(file)
			scanner := bufio.NewScanner(img.file)
			for scanner.Scan() {
				text := scanner.Bytes()
				text = append(text, '\n')
				_, err := writer.Write(append(text)[:])
				if err != nil {
					log.Println(err)
					return
				}

			}

			err = writer.Flush()
			if err != nil {
				log.Println("error while uploading to s3")
			}
		}()

	}
}

package assets

import (
	"bytes"
	"github.com/nfnt/resize"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"path"
	"sync"
	"time"
)

type OptFunc func(image *Image) error

type Image struct {
	ID        int64     `json:"id"`
	EventID   int64     `json:"eventID"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	Filename  *string   `json:"-"`
	ext       string
	file      io.Reader
}

func WithWidthAndHeight(w, h uint) OptFunc {
	return func(image *Image) error {
		if w == 0 && h == 0 {
			return nil
		}

		err := image.ResizeImage(w, h)
		if err != nil {
			return err
		}
		return nil
	}
}

func NewImage(filename string, file io.Reader, opts ...OptFunc) (Image, error) {

	image := Image{
		Filename: &filename,
		file:     file,
		ext:      path.Ext(filename),
	}

	wg := &sync.WaitGroup{}

	for _, fn := range opts {
		wg.Add(1)
		fn := fn
		go func() {
			defer wg.Done()
			err := fn(&image)
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
	wg.Wait()
	return image, nil
}
func (u *Image) SetFilename(filename string) {
	u.Filename = new(string)
	*u.Filename = filename
}
func (i *Image) ResizeImage(w uint, h uint) error {

	if i.ext == ".jpg" || i.ext == ".jpeg" {

		image, err := jpeg.Decode(i.file)
		if err != nil {
			return err
		}

		var newImage = resize.Resize(w, h, image, resize.Lanczos3)

		newFile := bytes.NewBuffer(make([]byte, 0))

		err = jpeg.Encode(newFile, newImage, nil)
		if err != nil {
			return err
		}

		i.file = newFile

	} else if i.ext == ".png" {

		image, err := png.Decode(i.file)
		if err != nil {
			return err
		}

		newImage := resize.Resize(w, h, image, resize.Lanczos3)

		newFile := bytes.NewBuffer(make([]byte, 0))

		err = png.Encode(newFile, newImage)
		if err != nil {
			return err
		}

		i.file = newFile

	}

	return nil
}

var (
	GetContentType = map[string]string{
		".webp": "image/webp",
		".jpeg": "image/jpeg",
		".jpg":  "image/jpeg",
		".png":  "image/png",
		".pdf":  "application/pdf",
	}
)

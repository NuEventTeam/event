package cdn

import (
	"bytes"
	"fmt"
	"github.com/NuEventTeam/events/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

type Content struct {
	FieldName string
	Filename  string
	Size      int64
	Reader    io.Reader
}

type CdnSvc struct {
	BaseURL string
}

func New(url string) *CdnSvc {
	return &CdnSvc{BaseURL: url}
}

type UploadImageResponse struct {
	Ok   bool     `json:"ok"`
	Path []string `json:"paths"`
}

func (c *CdnSvc) Upload(namespace any, files ...Content) error {
	postUrl, err := url.Parse(c.BaseURL + "/upload")
	if err != nil {
		return err
	}

	query := postUrl.Query()
	query.Add("namespace", fmt.Sprintf("%v", namespace))
	query.Add("width", "500")
	query.Add("height", "500")
	postUrl.RawQuery = query.Encode()

	data, err := sendFormDataRequest(postUrl.String(), files...)
	if err != nil {
		return err
	}

	resp := &UploadImageResponse{}

	err = sonic.ConfigFastest.Unmarshal(data, resp)
	if err != nil {
		return fmt.Errorf("something went wrong during upload")
	}

	return nil
}

func (c *CdnSvc) Delete(urls ...string) {

	for _, u := range urls {
		u := u
		go func() {
			request := pkg.Request{
				URL:    fmt.Sprintf("%s/delete/%s", c.BaseURL, u),
				Method: fiber.MethodDelete,
			}

			request.Send()
		}()
	}
}

func sendFormDataRequest(url string, files ...Content) ([]byte, error) {
	var (
		buf = new(bytes.Buffer)
		w   = multipart.NewWriter(buf)
	)

	for _, f := range files {
		if f.Reader == nil {
			err := w.WriteField(f.FieldName, f.Filename)
			if err != nil {
				return nil, err
			}
			continue
		}

		part, err := w.CreateFormFile(f.FieldName, filepath.Base(f.Filename))
		_, err = io.CopyN(part, f.Reader, f.Size)
		if err != nil {
			return nil, err
		}

	}
	w.Close()

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return []byte{}, err
	}

	req.Header.Add("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	log.Println("before request", time.Now())
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	defer res.Body.Close()

	cnt, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	log.Println("finished", time.Now())
	return cnt, nil
}

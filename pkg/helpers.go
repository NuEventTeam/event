package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

func Error(ctx *fiber.Ctx, status int, msg string, err ...error) error {

	log.Printf("%+v", err)

	return ctx.
		Status(status).
		JSON(fiber.Map{
			"ok": false,
			"data": fiber.Map{
				"message": msg,
			},
		})

}

func Success(ctx *fiber.Ctx, data interface{}) error {
	return ctx.
		Status(fiber.StatusOK).
		JSON(fiber.Map{
			"ok":   true,
			"data": data,
		})
}

type Request struct {
	URL    string
	Header map[string]string
	Method string
	Data   interface{}
}

func (r Request) Send() ([]byte, error) {
	js, err := sonic.ConfigFastest.Marshal(r.Data)
	if err != nil {
		return nil, err
	}

	agent := fiber.AcquireAgent()
	agent.Request().Header.SetMethod(r.Method)
	for key, val := range r.Header {
		agent.Set(key, val)
	}

	agent.Request().SetRequestURI(r.URL)
	agent.Body(js)

	err = agent.Parse()

	if err != nil {
		return nil, err
	}
	_, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return body, nil
}

func ParseJWT(jwtStr string, secret string) (int64, error) {
	token, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Invalid sign method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, jwt.ErrTokenExpired
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}

	userID := int64(claims["user_id"].(float64))

	return userID, nil
}

type Content struct {
	Fname  string
	Ftype  string
	Size   int64
	Reader io.Reader
}

func SendPostRequest(url string, files ...Content) ([]byte, error) {

	var (
		buf = new(bytes.Buffer)
		w   = multipart.NewWriter(buf)
	)

	for _, f := range files {
		if f.Reader == nil {
			err := w.WriteField(f.Fname, f.Ftype)
			if err != nil {
				return nil, err
			}
			continue
		}

		part, err := w.CreateFormFile(f.Fname, filepath.Base(f.Ftype))
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

type UploadImageResponse struct {
	Ok   bool     `json:"ok"`
	Path []string `json:"paths"`
}

func UploadImages(userId int64, imgs ...Content) ([]string, error) {
	postUrl, err := url.Parse(CDNBaseUrl + "/upload")
	if err != nil {
		return nil, err
	}

	query := postUrl.Query()
	query.Add("namespace", fmt.Sprintf("%d", userId))
	query.Add("width", "500")
	query.Add("height", "500")
	postUrl.RawQuery = query.Encode()

	data, err := SendPostRequest(postUrl.String(), imgs...)
	if err != nil {
		return nil, err
	}

	resp := &UploadImageResponse{}

	err = sonic.ConfigFastest.Unmarshal(data, resp)
	if err != nil {
		return nil, fmt.Errorf("something went wrong during upload")
	}

	return resp.Path, nil

}

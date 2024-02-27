package pkg

import (
	"errors"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"log"
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

	userID := int64(claims["userId"].(float64))

	return userID, nil
}

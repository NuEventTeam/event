package auth

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"time"
)

var (
	ErrPhoneExists = fmt.Errorf("phone already exists")
)

type RegisterRequest struct {
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	Token           string `json:"token"`
	ConfirmPassword string `json:"confirm_password"`
}

type RegisterResponse struct {
	AuthToken AuthToken `json:"tokens"`
}

func (a *Auth) RegisterHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request RegisterRequest

		if err := ctx.BodyParser(&request); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, MsgCannotParseJSON, err)
		}

		if request.ConfirmPassword != request.Password {
			return pkg.Error(ctx, fiber.StatusBadRequest, MsgConfirmPasswordNotSame)
		}

		if violation := ValidateLength(3, request.Password); violation != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, violation.Message, violation)
		}

		if violation := ValidatePhoneNumber(request.Phone); violation != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, violation.Message, violation)
		}

		token, err := a.VerifyToken(ctx.Context(),
			models.Token{Token: request.Token,
				Phone: &request.Phone,
				Type:  TokenTypeRegister,
			})
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		if token == nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "token expired")
		}

		if *token.Phone != request.Phone {
			return pkg.Error(ctx, fiber.StatusBadRequest, "incorrect phone number", fmt.Errorf("phone number are not the same"))
		}

		user := models.User{
			Phone:    request.Phone,
			Password: request.Password,
		}

		userID, err := a.CreateUser(ctx.Context(), user)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		refreshToken := models.Token{
			UserId: &userID,
			Token:  ulid.Make().String(),
			Type:   TokenTypeRefresh,

			Duration: 7 * 24 * time.Hour,
		}

		if val := ctx.Get("User-Agent", ""); val != "" {
			refreshToken.UserAgent = &val
		}

		err = a.CreateToken(ctx.Context(), refreshToken)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		accessToken, err := a.GetJWT(userID, refreshToken.UserAgent)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}

		response := RegisterResponse{
			AuthToken: AuthToken{
				AccessToken:  accessToken,
				RefreshToken: refreshToken.Token,
			},
		}

		return pkg.Success(ctx, response)
	}
}

func (a *Auth) CreateUser(ctx context.Context, user models.User) (int64, error) {

	exists, err := a.db.PhoneExists(ctx, a.db.GetDb(), user.Phone)
	if err != nil {
		return 0, err
	}

	if exists {
		return 0, ErrPhoneExists
	}

	hash, err := hashPassword(user.Password)
	if err != nil {
		return 0, err
	}
	user.Hash = hash

	userID, err := a.db.CreateUser(ctx, a.db.GetDb(), user)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

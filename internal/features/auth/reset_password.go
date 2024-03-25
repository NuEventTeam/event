package auth

import (
	"context"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

type ResetPasswordRequest struct {
	Password        string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
	Token           string `json:"token"`
}

func (a *Auth) ResetPasswordHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request ResetPasswordRequest

		if err := ctx.BodyParser(&request); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, MsgCannotParseJSON)
		}

		if request.Password != request.ConfirmPassword {
			return pkg.Error(ctx, fiber.StatusBadRequest, MsgConfirmPasswordNotSame)
		}

		token, err := a.VerifyToken(ctx.Context(), models.Token{Token: request.Token, Type: TokenTypeReset})
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}

		if token == nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "token expired")
		}

		user, err := a.db.GetUser(ctx.Context(), a.db.GetDb(), database.GetUserParams{
			UserID: token.UserId,
		})
		if user == nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "something went wrong", err)
		}

		if err := a.UpdatePassword(ctx.Context(), request.Password, user.ID); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "something went wrong", err)
		}

		err = database.DeleteToken(ctx.Context(), a.db.GetDb(), *token)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "something went wrong", err)
		}

		return pkg.Success(ctx, nil)
	}
}
func (a *Auth) UpdatePassword(ctx context.Context, password string, userID int64) error {
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	return a.db.UpdateUser(ctx, a.db.GetDb(), nil, &hash, userID)
}

package auth

import (
	"context"
	"errors"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type RefreshTokenRequest struct {
	Token string `json:"token"`
}

type RefreshTokenResponse struct {
	Token string `json:"token"`
}

func (a *Auth) RefreshTokenHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request RefreshTokenRequest

		if err := ctx.BodyParser(&request); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, MsgCannotParseJSON)
		}
		token, err := a.VerifyToken(ctx.Context(), models.Token{Token: request.Token, Type: TokenTypeRefresh})
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}
		if token == nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "token expired")
		}
		accessToken, err := a.GetJWT(*token.UserId, token.UserAgent)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}

		return pkg.Success(ctx, fiber.Map{"token": accessToken})
	}
}

func (a *Auth) VerifyToken(ctx context.Context, token models.Token) (*models.Token, error) {

	t, err := database.GetToken(ctx, a.db.GetDb(), token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if token.Type != TokenTypeRefresh {
		err = database.DeleteToken(ctx, a.db.GetDb(), t)
		if err != nil {
			return nil, err
		}
	}
	return &t, nil
}

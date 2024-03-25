package auth

import (
	"context"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func (a *Auth) LogoutHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		userId := ctx.Locals("userId").(int64)
		var agent *string
		if ctx.Locals("userAgent").(string) != "" {
			t := ctx.Locals("userAgent").(string)
			agent = &t
		}
		err := a.Logout(ctx.Context(), models.Token{UserId: &userId, Type: TokenTypeRefresh, UserAgent: agent})

		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "logout failed", err)
		}

		return pkg.Success(ctx, nil)
	}
}

func (a *Auth) Logout(ctx context.Context, token models.Token) error {
	return database.DeleteToken(ctx, a.db.GetDb(), token)
}

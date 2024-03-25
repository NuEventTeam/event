package user

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func (u User) GetByUsername() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		username := ctx.Params("username")
		user, err := u.GetUserByUsername(ctx.Context(), username)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, "/get/", *user.ProfileImage)

		user.ProfileImage = &profileImgUrl

		return pkg.Success(ctx, user)
	}

}

func (e *User) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	profile, err := database.GetUser(ctx, e.db.GetDb(), database.GetUserArgss{Username: &username})
	if err != nil {
		return models.User{}, err
	}

	preferences, err := database.GetUserPreferences(ctx, e.db.GetDb(), profile.UserID)
	if err != nil {
		return models.User{}, err
	}

	profile.Preferences = preferences

	return profile, nil
}

func (u User) checkUsername(ctx *fiber.Ctx) error {
	username := ctx.Params("username", "")
	if username == "" {
		return pkg.Error(ctx, fiber.StatusBadRequest, "empty username", fmt.Errorf("empty username"))
	}

	exists, err := u.CheckUsername(ctx.Context(), username)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}

	if exists {
		return ctx.SendStatus(fiber.StatusFound)
	}

	return ctx.SendStatus(fiber.StatusOK)

}

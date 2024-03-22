package event

import (
	"context"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func (e Event) GetAllCategoriesHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		categories := []models.Category{}
		if ctx.QueryInt("all") == 1 {
			cats, err := e.getAllCategories(ctx.Context(), nil)
			if err != nil {
				return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
			}
			categories = cats
		}
		return pkg.Success(ctx, fiber.Map{"categories": categories})
	}
}

func (e Event) getAllCategories(ctx context.Context, ids []int64) ([]models.Category, error) {
	categories, err := database.GetCategories(ctx, e.db.GetDb(), database.GetCategoriesParams{IDs: ids})
	return categories, err
}

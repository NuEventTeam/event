package event_service

import (
	"context"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
)

func (e *EventSvc) GetCategoriesByID(ctx context.Context, ids []int64) ([]models.Category, error) {
	categories, err := database.GetCategories(ctx, e.db.DB, database.GetCategoriesParams{IDs: ids})
	return categories, err
}

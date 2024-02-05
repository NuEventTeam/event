package database

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/models"
)

type GetCategoriesParams struct {
	Title *string
	IDs   []int64
}

func GetCategories(ctx context.Context, db DBTX, params GetCategoriesParams) ([]models.Category, error) {
	query := qb.Select("id", "name").From("categories").Where(sq.NotEq{"deleted_at": nil})

	if params.Title != nil {
		query = query.Where("name LIKE ?", fmt.Sprint("%", *params.Title, "%"))
	}

	if params.IDs != nil {
		qb.Where(sq.Eq{"id": params.IDs})
	}

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var res []models.Category

	rows, err := db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var each models.Category

		err := rows.Scan(&each.ID, &each.Name)
		if err != nil {
			return nil, err
		}

		res = append(res, each)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return res, nil
}

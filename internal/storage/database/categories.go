package database

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/models"
	"log"
)

type GetCategoriesParams struct {
	Title *string
	IDs   []int64
}

func GetCategories(ctx context.Context, db DBTX, params GetCategoriesParams) ([]models.Category, error) {
	query := qb.Select("id", "name").From("categories").Where(sq.Eq{"deleted_at": nil})

	if params.Title != nil {
		query = query.Where("name LIKE ?", fmt.Sprint("%", *params.Title, "%"))
	}

	if params.IDs != nil {
		query = query.Where(sq.Eq{"id": params.IDs})
	}

	stmt, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	log.Println(stmt)
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

package database

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/models"
	"log"
)

func CreateUser(ctx context.Context, db DBTX, user models.User) (int64, error) {
	query := qb.Insert("users").
		Columns("user_id", "phone", "username", "lastname", "firstname", "profile_image", "birthdate").
		Values(user.UserID, user.Phone, user.Username, user.Lastname, user.Firstname, user.ProfileImage, user.DateOfBirth).
		Suffix("returning id")

	stmt, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var id int64

	err = db.QueryRow(ctx, stmt, args...).Scan(&id)

	return id, err
}

func CheckUsername(ctx context.Context, db DBTX, username string) (bool, error) {
	query := qb.Select("count(*)").
		From("users").
		Where(sq.Eq{"username": username}).
		Limit(1)

	stmt, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var count int

	err = db.QueryRow(ctx, stmt, args...).Scan(&count)

	if err != nil {
		return false, err
	}

	return count != 0, nil
}

func UpdateUser(ctx context.Context, db DBTX, userID int64, params map[string]interface{}) error {
	query := qb.Update("users").
		SetMap(params).
		Where(sq.Eq{"user_id": userID})

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err

}

type GetUserArgss struct {
	UserID   *int64
	Username *string
}

func GetUser(ctx context.Context, db DBTX, args GetUserArgss) (models.User, error) {
	query := qb.Select("id", "user_id", "phone", "username", "lastname", "firstname", "birthdate", "profile_image").
		From("users").
		Where(sq.Eq{"deleted_at": nil})

	if args.UserID != nil {
		query = query.Where(sq.Eq{"user_id": *args.UserID})
	}
	if args.Username != nil {
		query = query.Where(sq.Eq{"username": *args.Username})
	}

	stmt, params, err := query.ToSql()
	if err != nil {
		return models.User{}, nil
	}

	var user models.User

	err = db.QueryRow(ctx, stmt, params...).Scan(&user.ID, &user.UserID, &user.Phone, &user.Username, &user.Lastname,
		&user.Firstname, &user.DateOfBirth, &user.ProfileImage)

	return user, err

}

func GetUserPreferences(ctx context.Context, db DBTX, userID int64) ([]models.Category, error) {
	query := qb.Select("category_id", "name").
		From("user_preferences").
		InnerJoin("categories on user_preferences.category_id = categories.id").
		Where(sq.Eq{"categories.deleted_at": nil}).
		Where(sq.Eq{"user_id": userID})

	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	log.Println(stmt)
	var res []models.Category

	rows, err := db.Query(ctx, stmt, params...)
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

func AddUserPreference(ctx context.Context, db DBTX, userID int64, category ...models.Category) error {

	query := qb.Insert("user_preferences").
		Columns("user_id", "category_id")

	for i := range category {
		query = query.Values(userID, category[i].ID)
	}

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func RemoveUserPreference(ctx context.Context, db DBTX, userID int64, categoryID int64) error {

	query := qb.Delete("user_preferences").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"category_id": categoryID})

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

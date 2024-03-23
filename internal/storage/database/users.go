package database

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/jackc/pgx/v5"
)

func CreateUser(ctx context.Context, db DBTX, user models.User) (int64, error) {
	query := qb.Insert("users").
		Columns("user_id", "phone", "username", "lastname", "firstname", "profile_image", "birthdate").
		Values(user.UserID, user.Phone, user.Username, user.Lastname, user.Firstname, user.ProfileImage, user.BirthDate).
		Suffix("returning id")

	stmt, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var id int64

	err = db.QueryRow(ctx, stmt, args...).Scan(&id)

	return id, err
}

func CheckUserExists(ctx context.Context, db DBTX, userID int64) (bool, error) {
	query := qb.Select("count(*)").
		From("users").
		Where(sq.Eq{"user_id": userID}).
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

	return count > 0, nil
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

	return count > 0, nil
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
		&user.Firstname, &user.BirthDate, &user.ProfileImage)

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
	if len(category) == 0 {
		return nil
	}
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

func AddUserFollower(ctx context.Context, db DBTX, userId, followerId int64) error {
	query := qb.Insert("user_followers").
		Columns("user_id", "follower_id").
		Values(userId, followerId)

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func RemoveUserFollower(ctx context.Context, db DBTX, userId, followerId int64) error {
	query := qb.Delete("user_followers").
		Where(sq.Eq{"user_id": userId}).
		Where(sq.Eq{"follower_id": followerId})

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func BanUserFollower(ctx context.Context, db DBTX, userId, followerId int64) error {
	query := qb.Insert("banned_user_followers").
		Columns("user_id", "follower_id").
		Values(userId, followerId)

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func UpdateUserFollowerCount(ctx context.Context, db DBTX, userId, by int64) error {
	query := qb.Update("users").
		Set("follower_count", fmt.Sprintf("follower_count %d", by)).
		Where(sq.Eq{"user_id": userId})

	stmt, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, args...)
	return err
}

func (d *Database) PhoneExists(ctx context.Context, db DBTX, phone string) (bool, error) {

	query := qb.Select("count(*)").
		From("users").
		Where(sq.Eq{"phone": phone})

	stmt, params, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var count int64

	err = db.QueryRow(ctx, stmt, params...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count != 0, nil
}

func (d *Database) CreateUser(ctx context.Context, db DBTX, user models.User) (int64, error) {
	query := qb.Insert("users").
		Columns("phone", "password").
		Values(user.Phone, user.Hash).Suffix("returning id")

	stmt, params, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var id int64

	err = db.QueryRow(ctx, stmt, params...).Scan(&id)

	return id, err
}

type GetUserParams struct {
	Phone  *string
	UserID *int64
}

func (d *Database) GetUser(ctx context.Context, db DBTX, args GetUserParams) (*models.User, error) {
	query := qb.Select("id", "password").
		From("users").
		Where(sq.Eq{"deleted_at": nil})

	if args.Phone != nil {
		query = query.Where(sq.Eq{"phone": args.Phone})
	}

	if args.UserID != nil {
		query = query.Where(sq.Eq{"id": args.UserID})
	}

	stmt, params, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	var user models.User

	err = db.QueryRow(ctx, stmt, params...).Scan(&user.ID, &user.Hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, err
}

func (d *Database) UpdateUser(ctx context.Context, db DBTX, phone, password *string, userID int64) error {
	query := qb.Update("users")

	if phone != nil {
		query = query.Set("phone", phone)
	}

	if password != nil {
		query = query.Set("password", password)
	}

	query = query.Where(sq.Eq{"id": userID})
	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err
}

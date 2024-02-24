package users

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
)

type UserService struct {
	db    *database.Database
	cache *cache.Cache
}

func NewEventSvc(db *database.Database, cache *cache.Cache) *UserService {
	return &UserService{
		db:    db,
		cache: cache,
	}
}
func (e *UserService) GetUserById(ctx context.Context, userId int64) (models.User, error) {
	profile, err := database.GetUser(ctx, e.db.GetDb(), database.GetUserArgss{UserID: &userId})
	if err != nil {
		return models.User{}, err
	}

	preferences, err := database.GetUserPreferences(ctx, e.db.GetDb(), userId)
	if err != nil {
		return models.User{}, err
	}

	profile.Preferences = preferences

	return profile, nil
}
func (e *UserService) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
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

func (e *UserService) CreateUser(ctx context.Context, user models.User) error {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = database.CreateUser(ctx, tx, user)
	if err != nil {
		return err
	}

	err = database.AddUserPreference(ctx, tx, user.UserID, user.Preferences...)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (e *UserService) ChangeUserProfile(ctx context.Context, userId int64, params map[string]interface{}) error {
	err := database.UpdateUser(ctx, e.db.GetDb(), userId, params)
	return err
}

func (e *UserService) CheckUsername(ctx context.Context, username string) (bool, error) {
	exists, err := database.CheckUsername(ctx, e.db.GetDb(), username)
	return exists, err
}
func (e *UserService) CheckUserID(ctx context.Context, userID int64) (bool, error) {
	exists, err := database.CheckUserExists(ctx, e.db.GetDb(), userID)
	return exists, err
}

func (e *UserService) AddUserPreference(ctx context.Context, userID int64, category []models.Category) error {
	err := database.AddUserPreference(ctx, e.db.GetDb(), userID, category...)
	if err != nil {
		return err
	}
	return err
}
func (e *UserService) RemoveUserPreference(ctx context.Context, userID int64, categoryID int64) error {
	err := database.RemoveUserPreference(ctx, e.db.GetDb(), userID, categoryID)
	if err != nil {
		return err
	}
	return err
}

func (e *UserService) ChangeUsername(ctx context.Context, userId int64, username string) error {
	exists, err := database.CheckUsername(ctx, e.db.GetDb(), username)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("username exists")
	}

	updateParams := map[string]interface{}{
		"username": username,
	}

	err = database.UpdateUser(ctx, e.db.GetDb(), userId, updateParams)

	return err
}

func (e *UserService) GetCategoriesByID(ctx context.Context, ids []int64) ([]models.Category, error) {
	categories, err := database.GetCategories(ctx, e.db.GetDb(), database.GetCategoriesParams{IDs: ids})
	return categories, err
}

func (e *UserService) AddFollower(ctx context.Context, userId, followerId int64) error {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = database.AddUserFollower(ctx, tx, userId, followerId)
	if err != nil {
		return err
	}
	err = database.UpdateUserFollowerCount(ctx, tx, userId, 1)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (e *UserService) RemoveFollower(ctx context.Context, userId, followerId int64) error {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = database.RemoveUserFollower(ctx, tx, userId, followerId)
	if err != nil {
		return err
	}
	err = database.UpdateUserFollowerCount(ctx, tx, userId, -1)
	if err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (e *UserService) BanFollower(ctx context.Context, userId, followerId int64) error {
	tx, err := e.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	err = database.BanUserFollower(ctx, tx, userId, followerId)
	if err != nil {
		return err
	}
	//TODO check if exist then remove from follower and deacrease followers count
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

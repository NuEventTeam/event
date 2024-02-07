package event_service

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/jackc/pgx/v5"
	"log"
)

func (e *EventSvc) GetUserById(ctx context.Context, userId int64) (models.User, error) {
	profile, err := database.GetUser(ctx, e.db.DB, database.GetUserArgss{UserID: &userId})
	if err != nil {
		return models.User{}, err
	}

	preferences, err := database.GetUserPreferences(ctx, e.db.DB, userId)
	if err != nil {
		return models.User{}, err
	}

	profile.Preferences = preferences

	return profile, nil
}
func (e *EventSvc) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	profile, err := database.GetUser(ctx, e.db.DB, database.GetUserArgss{Username: &username})
	if err != nil {
		return models.User{}, err
	}
	log.Println(profile.UserID)
	preferences, err := database.GetUserPreferences(ctx, e.db.DB, profile.UserID)
	if err != nil {
		return models.User{}, err
	}

	profile.Preferences = preferences

	return profile, nil
}

func (e *EventSvc) CreateUser(ctx context.Context, user models.User) error {
	tx, err := e.db.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadUncommitted})
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

func (e *EventSvc) ChangeUserProfile(ctx context.Context, userId int64, params map[string]interface{}) error {
	err := database.UpdateUser(ctx, e.db.DB, userId, params)
	return err
}

func (e *EventSvc) CheckUsername(ctx context.Context, username string) (bool, error) {
	exists, err := database.CheckUsername(ctx, e.db.DB, username)
	return exists, err
}

func (e *EventSvc) AddUserPreference(ctx context.Context, userID int64, category []models.Category) error {
	err := database.AddUserPreference(ctx, e.db.DB, userID, category...)
	if err != nil {
		return err
	}
	return err
}
func (e *EventSvc) RemoveUserPreference(ctx context.Context, userID int64, categoryID int64) error {
	err := database.RemoveUserPreference(ctx, e.db.DB, userID, categoryID)
	if err != nil {
		return err
	}
	return err
}

func (e *EventSvc) ChangeUsername(ctx context.Context, userId int64, username string) error {
	exists, err := database.CheckUsername(ctx, e.db.DB, username)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("username exists")
	}

	updateParams := map[string]interface{}{
		"username": username,
	}

	err = database.UpdateUser(ctx, e.db.DB, userId, updateParams)

	return err
}

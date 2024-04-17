package user

import (
	"context"
	"fmt"
	user_follow "github.com/NuEventTeam/events/internal/features/user/follow"
	user_profile "github.com/NuEventTeam/events/internal/features/user/profile"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
)

func (u *User) GetByUsername() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		username := ctx.Params("username")
		user, err := u.GetUserByUsername(ctx.Context(), username)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}
		if user.ProfileImage != nil {
			profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, *user.ProfileImage)
			user.ProfileImage = &profileImgUrl
		}
		followedEventsMap, err := user_profile.GetFollowedEvents(ctx.Context(), u.db.GetDb(), user.UserID, 0)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		var followedEvents []user_profile.FollowedEvent
		for _, val := range followedEventsMap {
			followedEvents = append(followedEvents, val)
		}

		likedEventsMap, err := user_profile.GetLikedEvents(ctx.Context(), u.db.GetDb(), user.UserID, 0)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		var likedEvents []user_profile.FollowedEvent

		for _, val := range likedEventsMap {
			likedEvents = append(likedEvents, val)
		}

		pastEventsMap, err := user_profile.GetOldEvents(ctx.Context(), u.db.GetDb(), user.UserID, 0)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		var pastEvents []user_profile.FollowedEvent
		for _, val := range pastEventsMap {
			pastEvents = append(pastEvents, val)
		}

		ownEventsMap, err := user_profile.GetOwnEvents(ctx.Context(), u.db.GetDb(), user.UserID, 9)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		var ownEvents []user_profile.FollowedEvent
		for _, val := range ownEventsMap {
			ownEvents = append(ownEvents, val)
		}

		followedUsers, err := user_follow.GetFollowed(ctx.Context(), u.db.GetDb(), user.UserID)
		if err != nil {
			if err != nil {
				return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
			}
		}
		response := fiber.Map{
			"user":          user,
			"own":           ownEvents,
			"followed_user": followedUsers,
		}

		userId := ctx.Locals("userId")
		if userId != nil && userId.(int64) == user.UserID {
			response["events"] = fiber.Map{
				"followed":  followedEvents,
				"favourite": followedEvents,
				"past":      followedEvents}
		}

		return pkg.Success(ctx, response)
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

func (u *User) GetOwnUserProfile() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("userId").(int64)
		user, err := database.GetUser(ctx.Context(), u.db.GetDb(), database.GetUserArgss{UserID: &userId})
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}
		if user.ProfileImage != nil {
			profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, *user.ProfileImage)
			user.ProfileImage = &profileImgUrl
		}
		preferences, err := database.GetUserPreferences(ctx.Context(), u.db.GetDb(), user.UserID)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}
		user.Preferences = preferences

		followedEventsMap, err := user_profile.GetFollowedEvents(ctx.Context(), u.db.GetDb(), user.UserID, 0)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		var followedEvents []user_profile.FollowedEvent
		for _, val := range followedEventsMap {
			followedEvents = append(followedEvents, val)
		}

		likedEventsMap, err := user_profile.GetLikedEvents(ctx.Context(), u.db.GetDb(), user.UserID, 0)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		var likedEvents []user_profile.FollowedEvent

		for _, val := range likedEventsMap {
			likedEvents = append(likedEvents, val)
		}

		pastEventsMap, err := user_profile.GetOldEvents(ctx.Context(), u.db.GetDb(), user.UserID, 0)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		var pastEvents []user_profile.FollowedEvent
		for _, val := range pastEventsMap {
			pastEvents = append(pastEvents, val)
		}

		ownEventsMap, err := user_profile.GetOwnEvents(ctx.Context(), u.db.GetDb(), user.UserID, 9)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		var ownEvents []user_profile.FollowedEvent
		for _, val := range ownEventsMap {
			ownEvents = append(ownEvents, val)
		}

		followedUsers, err := user_follow.GetFollowed(ctx.Context(), u.db.GetDb(), user.UserID)
		if err != nil {
			if err != nil {
				return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
			}
		}
		response := fiber.Map{
			"user":          user,
			"own":           ownEvents,
			"followed_user": followedUsers,
			"events": fiber.Map{
				"followed":  followedEvents,
				"favourite": followedEvents,
				"past":      followedEvents,
			},
		}

		return pkg.Success(ctx, response)
	}

}

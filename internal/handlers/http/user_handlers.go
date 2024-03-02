package http

import (
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/services/cdn"
	"github.com/NuEventTeam/events/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"log"
	"path"
	"strconv"
	"time"
)

func (h *Handler) SetUpUserRoutes(router *fiber.App) {
	apiV1 := router.Group("/api/v1")

	apiV1.Get("/categories", h.getCategories)

	apiV1.Post("/create/user", MustAuth(h.JWTSecret), h.createUser)

	apiV1.Post("/create/mobile/user", MustAuth(h.JWTSecret), h.createMobileUser)

	apiV1.Get("/user/:username", h.getUser)

	apiV1.Get("/check-username/:username", MustAuth(h.JWTSecret), h.checkUsername)

	apiV1.Post("/users/friendship/follow/:userId", MustAuth(h.JWTSecret), h.followUser)

	apiV1.Post("/users/friendship/unfollow/:userId", MustAuth(h.JWTSecret), h.unfollowUser)

}

func (h *Handler) followUser(ctx *fiber.Ctx) error {
	followerId := ctx.Locals("userId").(int64)
	userId, err := strconv.ParseInt(ctx.Params("userId"), 10, 64)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "invalid follower id", err)
	}

	err = h.userSvc.AddFollower(ctx.Context(), userId, followerId)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
	}
	return pkg.Success(ctx, nil)
}

func (h *Handler) unfollowUser(ctx *fiber.Ctx) error {
	followerId := ctx.Locals("userId").(int64)
	userId, err := strconv.ParseInt(ctx.Params("userId"), 10, 64)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "invalid follower id", err)
	}

	err = h.userSvc.RemoveFollower(ctx.Context(), userId, followerId)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
	}
	return pkg.Success(ctx, nil)
}

func (h *Handler) getUser(ctx *fiber.Ctx) error {
	username := ctx.Params("username")
	user, err := h.userSvc.GetUserByUsername(ctx.Context(), username)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
	}

	if user.ProfileImage != nil {
		profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, "/get/", *user.ProfileImage)
		user.ProfileImage = &profileImgUrl
	}

	return pkg.Success(ctx, user)
}

type MobileUser struct {
	UserID       int64   `json:"userID,omitempty"`
	Username     string  `json:"username,omitempty"`
	ProfileImage *string `json:"profileImage,omitempty"`
	BirthDate    string  `json:"birthDate,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	Firstname    string  `json:"firstname,omitempty"`
	Lastname     *string `json:"lastname,omitempty"`
	Preferences  []int64 `json:"preferences,omitempty"`
}

func (h *Handler) createMobileUser(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}

	userId := ctx.Locals("userId").(int64)

	var request MobileUser
	if len(form.Value["payload"]) == 0 {
		return pkg.Error(ctx, fiber.StatusBadRequest, "payload is missing", fmt.Errorf("payload is missing"))
	}

	err = sonic.ConfigFastest.Unmarshal([]byte(form.Value["payload"][0]), &request)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, "cannot read payload")
	}

	exists, err := h.userSvc.CheckUsername(ctx.Context(), request.Username)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}
	if exists {
		return pkg.Error(ctx, fiber.StatusBadRequest, "username exists")
	}

	exists, err = h.userSvc.CheckUserID(ctx.Context(), userId)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}
	if exists {
		return pkg.Error(ctx, fiber.StatusBadRequest, "user exists")
	}
	var imgs cdn.Content
	var profileImage *string
	log.Println(len(form.File["images"]))
	for _, f := range form.File["images"] {
		file, err := f.Open()
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "cannot open file", err)
		}
		filename := ulid.Make().String() + path.Ext(f.Filename)
		str := fmt.Sprint(pkg.UserNamespace, "/", userId, "/", filename)
		imgs = cdn.Content{
			FieldName: "files",
			Filename:  filename,
			Payload:   file,
			Size:      f.Size,
		}
		log.Println(imgs)
		profileImage = &str
		break
	}
	log.Println(imgs)
	err = h.cdnSvc.Upload(fmt.Sprintf(pkg.UserNamespace, "/", userId), imgs)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
	}

	request.UserID = userId
	request.ProfileImage = profileImage
	birthDate, err := time.Parse(time.DateOnly, request.BirthDate)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)

	}

	var categoryIds []models.Category
	for _, id := range request.Preferences {
		categoryIds = append(categoryIds, models.Category{ID: id})
	}

	if pkg.MinCategories > len(categoryIds) {
		return pkg.Error(ctx, fiber.StatusBadRequest, "must include 3 cats", fmt.Errorf("minimum of 3 cats"))

	}
	err = h.userSvc.CreateUser(ctx.Context(), models.User{
		UserID:       userId,
		Phone:        request.Phone,
		Username:     request.Username,
		Firstname:    request.Firstname,
		Lastname:     request.Lastname,
		ProfileImage: request.ProfileImage,
		BirthDate:    birthDate,
		Preferences:  categoryIds,
	})
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
	}

	return pkg.Success(ctx, fiber.Map{"username": request.Username})
}

func (h *Handler) createUser(ctx *fiber.Ctx) error {
	form, err := ctx.MultipartForm()
	if err != nil {
		return err
	}

	userId := ctx.Locals("userId").(int64)

	exists, err := h.userSvc.CheckUsername(ctx.Context(), form.Value["username"][0])
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}
	if exists {
		return pkg.Error(ctx, fiber.StatusBadRequest, "user exists")
	}

	exists, err = h.userSvc.CheckUserID(ctx.Context(), userId)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}
	if exists {
		return pkg.Error(ctx, fiber.StatusBadRequest, "username exists")
	}

	var categoryIds []models.Category
	for _, i := range form.Value["category_ids"] {
		id, _ := strconv.ParseInt(i, 10, 64)
		categoryIds = append(categoryIds, models.Category{ID: id})
	}
	if pkg.MinCategories > len(categoryIds) {
		return pkg.Error(ctx, fiber.StatusBadRequest, "must include 3 cats", fmt.Errorf("minimum of 3 cats"))

	}
	_, err = time.Parse(time.DateOnly, form.Value["birthdate"][0])
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}

	var imgs cdn.Content
	var profileImage *string
	for _, f := range form.File["images"] {
		file, err := f.Open()
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, "cannot open file", err)
		}
		filename := ulid.Make().String() + path.Ext(f.Filename)
		str := fmt.Sprint(pkg.UserNamespace, "/", userId, "/", filename)
		imgs = cdn.Content{
			FieldName: "files",
			Filename:  filename,
			Payload:   file,
			Size:      f.Size,
		}
		log.Println(imgs)
		profileImage = &str
		break
	}
	log.Println(imgs)
	err = h.cdnSvc.Upload(fmt.Sprintf(pkg.UserNamespace, "/", userId), imgs)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
	}

	birthDate, err := time.Parse(time.DateOnly, form.Value["birthdate"][0])
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)

	}
	err = h.userSvc.CreateUser(ctx.Context(), models.User{
		UserID:       userId,
		Username:     form.Value["username"][0],
		ProfileImage: profileImage,
		BirthDate:    birthDate,
		Phone:        form.Value["phone"][0],
		Firstname:    form.Value["firstname"][0],
		Lastname:     &form.Value["lastname"][0],
		Preferences:  categoryIds,
	})
	if err != nil {
		return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
	}

	return pkg.Success(ctx, fiber.Map{"username": form.Value["username"][0]})

}

func (h *Handler) checkUsername(ctx *fiber.Ctx) error {
	username := ctx.Params("username", "")
	if username == "" {
		return pkg.Error(ctx, fiber.StatusBadRequest, "empty username", fmt.Errorf("empty username"))
	}

	exists, err := h.userSvc.CheckUsername(ctx.Context(), username)
	if err != nil {
		return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
	}
	if exists {
		return ctx.SendStatus(fiber.StatusFound)
	}
	return ctx.SendStatus(fiber.StatusOK)

}

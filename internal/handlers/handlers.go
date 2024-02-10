package handlers

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	event_service "github.com/NuEventTeam/events/internal/services/event"
	"github.com/NuEventTeam/events/internal/storage/cache"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/NuEventTeam/protos/gen/go/event"

	"time"
)

type GRPCHandler struct {
	port     string
	timeout  time.Duration
	eventSvc *event_service.EventSvc
	database *database.Database
	cache    *cache.Cache
	event.UnimplementedEventServiceServer
	event.UnimplementedCategoriesServiceServer
	event.UnimplementedUserServiceServer
}

func NewGRPCHandler(
	database *database.Database,
	cache *cache.Cache,
	eventSvc *event_service.EventSvc,
) *GRPCHandler {

	return &GRPCHandler{
		database: database,
		cache:    cache,
		eventSvc: eventSvc,
	}
}

func (h *GRPCHandler) CreateEvent(ctx context.Context, request *event.CreateEventRequest) (*event.CreateEventResponse, error) {

	images := []models.Image{}

	for _, i := range request.Images {
		images = append(images, models.Image{
			Url: i,
		})
	}

	locations := []models.Location{}

	for _, l := range request.Locations {
		startsAt, err := time.Parse(time.DateTime, l.StartAt)
		endsAt, err := time.Parse(time.DateTime, l.StartAt)
		if err != nil {
			return nil, err
		}
		locations = append(locations, models.Location{
			Address:   l.Address,
			Longitude: float64(l.Longitude),
			Latitude:  float64(l.Latitude),
			StartsAt:  startsAt,
			EndsAt:    endsAt,
			Seats:     l.Seats,
		})
	}

	managers := []models.Manager{{
		User: models.User{UserID: request.Author},
		Role: models.Role{
			Name:        pkg.AuthorTitle,
			Permissions: []int64{pkg.PermissionRead, pkg.PermissionVerify, pkg.PermissionVerify}},
	}}

	categories := []models.Category{}
	for _, c := range request.Categories {
		categories = append(categories, models.Category{
			ID: c,
		})
	}

	e := models.Event{
		Title:       request.Title,
		Description: request.Description,
		MaxAge:      request.AgeMax,
		MinAge:      request.AgeMin,
		Locations:   locations,
		Images:      images,
		Managers:    managers,
		Categories:  categories,
	}

	eventID, err := h.eventSvc.CreateEvent(ctx, e)
	if err != nil {
		return nil, err
	}

	return &event.CreateEventResponse{
		Ok:      true,
		Message: "success",
		EventID: eventID,
	}, nil
}

func (h *GRPCHandler) GetCategories(ctx context.Context, request *event.GetCategoriesRequest) (*event.GetCategoriesResponse, error) {
	cats, err := h.eventSvc.GetCategoriesByID(ctx, nil)
	if err != nil {
		return nil, err
	}

	categories := []*event.Category{}
	for _, c := range cats {
		categories = append(categories, &event.Category{
			Id:   c.ID,
			Name: c.Name,
		})
	}

	return &event.GetCategoriesResponse{
		Categories: categories,
		Ok:         true,
		Message:    "success",
	}, nil
}

func (h *GRPCHandler) CreateUser(ctx context.Context, request *event.CreateUserRequest) (*event.CreateUserResponse, error) {

	exist, err := h.eventSvc.CheckUsername(ctx, request.User.Username)
	if err != nil {
		return nil, err
	}

	if exist {
		return &event.CreateUserResponse{Ok: false, Message: "username exists"}, nil
	}

	dateOfBirth, err := time.Parse(time.DateOnly, request.User.BirthDate)
	if err != nil {
		return nil, err
	}
	var preferences []models.Category
	for _, val := range request.Categories {
		preferences = append(preferences, models.Category{ID: val.Id})
	}

	user := models.User{
		UserID:       request.User.UserID,
		Phone:        request.User.Phone,
		Username:     request.User.Username,
		Firstname:    request.User.Firstname,
		Lastname:     request.User.Lastname,
		ProfileImage: request.User.ProfileImage,
		DateOfBirth:  dateOfBirth,
		Preferences:  preferences,
	}

	err = h.eventSvc.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return &event.CreateUserResponse{Ok: true, Message: "success"}, nil
}

func (h *GRPCHandler) CheckUsername(ctx context.Context, request *event.CheckUsernameRequest) (*event.CheckUsernameResponse, error) {
	exists, err := h.eventSvc.CheckUsername(ctx, request.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return &event.CheckUsernameResponse{Ok: false, Message: "username exists"}, nil
	}

	return &event.CheckUsernameResponse{Ok: true}, nil
}
func (h *GRPCHandler) AddUserPreferences(ctx context.Context, request *event.AddUserPreferencesRequest) (*event.AddUserPreferencesResponse, error) {

	cats := make([]models.Category, len(request.CategoryIDs))

	for _, c := range request.CategoryIDs {
		cats = append(cats, models.Category{ID: c})
	}

	err := h.eventSvc.AddUserPreference(ctx, request.UserId, cats)
	if err != nil {
		return nil, err
	}

	return &event.AddUserPreferencesResponse{
		Ok:      true,
		Message: "success",
	}, nil

}

func (h *GRPCHandler) GetEventByID(ctx context.Context, request *event.GetEventByIDRequest) (*event.GetEventByIDResponse, error) {
	e, err := h.eventSvc.GetEventByID(ctx, request.EventID)
	if err != nil {
		return nil, err
	}

	var locations []*event.Location

	for _, l := range e.Locations {
		locations = append(locations, &event.Location{
			Id:        l.ID,
			EventId:   l.EventID,
			Address:   l.Address,
			Longitude: float32(l.Longitude),
			Latitude:  float32(l.Latitude),
			Seats:     l.Seats,
			StartAt:   l.StartsAt.Format(time.DateTime),
			EndAt:     l.EndsAt.Format(time.DateTime),
		})
	}

	var categories []*event.Category

	for _, c := range e.Categories {
		categories = append(categories, &event.Category{
			Id:   c.ID,
			Name: c.Name,
		})
	}

	var images []string

	for _, i := range e.Images {
		images = append(images, i.Url)
	}

	var managers []*event.Managers

	for _, m := range e.Managers {
		managers = append(managers, &event.Managers{
			UserID: m.User.UserID,
			Role: &event.Role{
				Id:          m.Role.ID,
				Name:        m.Role.Name,
				Permissions: nil,
			},
		})
	}

	response := &event.Event{
		Id:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		AgeMax:      e.MaxAge,
		AgeMin:      e.MinAge,
		Locations:   locations,
		Managers:    managers,
		Categories:  categories,
		Images:      images,
	}

	return &event.GetEventByIDResponse{
		Event:   response,
		Ok:      true,
		Message: "succes",
	}, nil
}

func (h *GRPCHandler) RemoveUserPreferences(ctx context.Context, request *event.RemoveUserPreferencesRequest) (*event.RemoveUserPreferencesResponse, error) {

	err := h.eventSvc.RemoveUserPreference(ctx, request.UserId, request.CategoryIDs)
	if err != nil {
		return nil, err
	}
	return &event.RemoveUserPreferencesResponse{
		Ok:      true,
		Message: "success",
	}, nil
}
func (h *GRPCHandler) EditUserProfile(ctx context.Context, request *event.EditUserProfileRequest) (*event.EditUserProfileResponse, error) {
	params := map[string]interface{}{}

	if request.Phone != nil {
		params["phone"] = *request.Phone
	}

	if request.Username != nil {
		exists, err := h.eventSvc.CheckUsername(ctx, *request.Username)
		if err != nil {
			return nil, err
		}
		if exists {
			return &event.EditUserProfileResponse{Ok: false, Message: "username exists"}, nil
		}
		params["username"] = *request.Username
	}

	if request.ProfileImage != nil {
		params["profile_image"] = *request.ProfileImage
	}

	err := h.eventSvc.ChangeUserProfile(ctx, request.UserID, params)
	if err != nil {
		return nil, err
	}
	return &event.EditUserProfileResponse{
		Ok:      true,
		Message: "success",
	}, nil
}

func (h *GRPCHandler) GetUser(ctx context.Context, request *event.GetUserRequest) (*event.User, error) {

	user, err := h.eventSvc.GetUserByUsername(ctx, request.Username)
	if err != nil {
		return nil, err
	}

	profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, "/get/", *user.ProfileImage)

	categories := []*event.Category{}

	for _, p := range user.Preferences {
		categories = append(categories, &event.Category{
			Id:   p.ID,
			Name: p.Name,
		})
	}
	return &event.User{
		UserID:       user.UserID,
		Username:     user.Username,
		ProfileImage: &profileImgUrl,
		BirthDate:    user.DateOfBirth.Format(time.DateOnly),
		Phone:        user.Phone,
		Firstname:    user.Firstname,
		Lastname:     user.Lastname,
		Preferences:  categories,
	}, nil
}

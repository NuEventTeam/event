package auth

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

var (
	ErrInvalidCredentials     = fmt.Errorf("invalid credentials")
	MsgConfirmPasswordNotSame = "passwords are not the same"
	MsgCannotParseJSON        = "invalid json"
)

type AuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AuthToken AuthToken   `json:"tokens"`
	UserID    int64       `json:"user_id"`
	User      models.User `json:"user"`
}

func (a *Auth) LoginHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request LoginRequest

		if err := ctx.BodyParser(&request); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, MsgCannotParseJSON, err)
		}

		userID, err := a.CheckUserCredentials(ctx.Context(), &request.Phone, nil, request.Password)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)

		}

		refreshToken := models.Token{
			UserId:   &userID,
			Token:    ulid.Make().String(),
			Type:     TokenTypeRefresh,
			Duration: 7 * 24 * time.Hour,
		}
		if val := ctx.Get("User-Agent", ""); val != "" {
			refreshToken.UserAgent = &val
		}

		err = a.CreateToken(ctx.Context(), refreshToken)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)

		}

		accessToken, err := a.GetJWT(userID, refreshToken.UserAgent)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)

		}

		user, err := database.GetUser(ctx.Context(), a.db.GetDb(), database.GetUserArgss{UserID: &userID})
		if err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, err.Error(), err)
		}
		if user.ProfileImage != nil {
			profileImgUrl := fmt.Sprint(pkg.CDNBaseUrl, "/get/", *user.ProfileImage)
			user.ProfileImage = &profileImgUrl
		}
		response := LoginResponse{
			AuthToken: AuthToken{
				AccessToken:  accessToken,
				RefreshToken: refreshToken.Token,
			},
			User:   user,
			UserID: userID,
		}

		return pkg.Success(ctx, response)
	}
}

func (a *Auth) CreateToken(ctx context.Context, token models.Token) error {
	tx, err := a.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	err = database.DeleteToken(ctx, tx, token)
	if err != nil {
		return err
	}
	err = database.CreateToken(ctx, tx, token)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (a *Auth) GetJWT(userID int64, userAgent *string) (string, error) {
	var (
		key = []byte(a.jwt.Secret)
	)
	log.Println(userID)
	expireTime := time.Now().Add(a.jwt.Expiry)
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["userId"] = userID
	claims["exp"] = expireTime.Unix()
	if userAgent != nil {
		claims["userAgent"] = userAgent
	}
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func (a *Auth) CheckUserCredentials(ctx context.Context, phone *string, userID *int64, password string) (int64, error) {
	user, err := a.db.GetUser(ctx, a.db.GetDb(), database.GetUserParams{
		Phone:  phone,
		UserID: userID,
	})
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, fmt.Errorf("something wring with password or phone")
	}

	ok := checkPasswordHash(password, user.Hash)
	if !ok {
		return 0, ErrInvalidCredentials
	}

	return user.ID, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

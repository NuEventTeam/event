package auth

import (
	"context"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/oklog/ulid/v2"
	"time"
)

type ConfirmOtpRequest struct {
	Code    string `json:"code"`
	Phone   string `json:"phone"`
	OtpType int32  `json:"otp_type"`
}

type ConfirmOtpResponse struct {
	Token string `json:"token"`
}

func (a *Auth) VerifyOtpHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request ConfirmOtpRequest

		if err := ctx.BodyParser(&request); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, MsgCannotParseJSON)
		}

		if violation := ValidatePhoneNumber(request.Phone); violation != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, violation.Message, violation)
		}
		otp := models.Otp{
			Phone:   request.Phone,
			Code:    request.Code,
			OtpType: request.OtpType,
		}

		ok, err := a.VerifyOtp(ctx.Context(), otp)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}

		if !ok {
			return pkg.Error(ctx, fiber.StatusBadRequest, "wrong code")
		}

		token := models.Token{
			Token:    ulid.Make().String(),
			Type:     OtpToTokenType[request.OtpType],
			Duration: time.Minute,
		}
		if request.OtpType == OtpTypeRegister {
			token.Phone = &request.Phone
		} else {
			user, err := a.db.GetUser(ctx.Context(), a.db.GetDb(), database.GetUserParams{
				Phone: &request.Phone,
			})
			if err != nil {
				return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
			}
			if user == nil {
				return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
			}
			token.UserId = &user.ID
		}

		err = a.CreateToken(ctx.Context(), token)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}

		return pkg.Success(ctx, fiber.Map{"token": token.Token})
	}
}

func (a *Auth) VerifyOtp(ctx context.Context, otp models.Otp) (bool, error) {
	code, err := database.GetOtp(ctx, a.db.GetDb(), otp)
	if err != nil {
		return false, err
	}

	if code != otp.Code {
		return false, nil
	}

	err = database.DeleteOtp(ctx, a.db.GetDb(), otp)
	if err != nil {
		return false, err
	}
	return true, nil

}

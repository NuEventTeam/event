package auth

import (
	"context"
	"fmt"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/NuEventTeam/events/internal/storage/database"
	"github.com/NuEventTeam/events/pkg"
	"github.com/gofiber/fiber/v2"
	"math/rand"
	"strconv"
	"time"
)

type SendOtpRequest struct {
	Phone   string `json:"phone"`
	OtpType int32  `json:"otp_type"`
}

type SendOtpResponse struct {
	Code string `json:"code"`
}

func (a *Auth) SendOTPHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request SendOtpRequest

		if err := ctx.BodyParser(&request); err != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, MsgCannotParseJSON, err)
		}

		if violation := ValidatePhoneNumber(request.Phone); violation != nil {
			return pkg.Error(ctx, fiber.StatusBadRequest, violation.Message, violation)
		}

		exist, err := a.db.PhoneExists(ctx.Context(), a.db.GetDb(), request.Phone)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}

		if exist && OtpTypeRegister == request.OtpType {
			return pkg.Error(ctx, fiber.StatusBadRequest, fmt.Errorf("user already exists").Error(), fmt.Errorf("user already exists"))
		} else if !exist && OtpTypeRegister != request.OtpType {
			return pkg.Error(ctx, fiber.StatusBadRequest, fmt.Errorf("user not exists").Error(), fmt.Errorf("user not exists"))
		}

		otp, err := a.SaveOtp(ctx.Context(), request.Phone, request.OtpType)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, err.Error(), err)
		}

		err = a.smsProvider.Send(ctx.Context(), request.Phone, otp.Code)
		if err != nil {
			return pkg.Error(ctx, fiber.StatusInternalServerError, "something went wrong", err)
		}

		return pkg.Success(ctx, fiber.Map{"code": otp.Code})

	}
}

func (a *Auth) SaveOtp(ctx context.Context, phone string, otpType int32) (models.Otp, error) {

	otp := models.Otp{
		Phone:    phone,
		Code:     generateOtp(),
		OtpType:  otpType,
		Duration: time.Minute,
	}

	code, err := database.GetOtp(ctx, a.db.GetDb(), otp)
	if err != nil {
		return otp, err
	}
	if code != "" {
		return otp, fmt.Errorf("code already sent")
	}

	err = database.CreateOtp(ctx, a.db.GetDb(), otp)
	if err != nil {
		return models.Otp{}, err
	}
	return otp, nil
}

func generateOtp() string {
	return strconv.Itoa(rand.Intn(8999) + 1000)
}

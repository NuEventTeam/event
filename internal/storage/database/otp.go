package database

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/NuEventTeam/events/internal/models"
	"github.com/jackc/pgx/v5"
	"time"
)

const OtpTable = "one_time_passwords"

func CreateOtp(ctx context.Context, db DBTX, otp models.Otp) error {
	query := qb.Insert(OtpTable).
		Columns("phone", "code", "type", "expires_at").
		Values(otp.Phone, otp.Code, otp.OtpType, time.Now().Add(otp.Duration))

	stmt, params, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)

	return err

}

func GetOtp(ctx context.Context, db DBTX, otp models.Otp) (string, error) {
	query := qb.Select("code").From(OtpTable).
		Where(sq.Eq{"phone": otp.Phone}).
		Where(sq.Eq{"type": otp.OtpType}).
		Where(sq.Gt{"expires_at": time.Now()})

	stmt, params, err := query.ToSql()
	if err != nil {
		return "", err
	}
	var code string
	err = db.QueryRow(ctx, stmt, params...).Scan(&code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return code, nil
}

func DeleteOtp(ctx context.Context, db DBTX, otp models.Otp) error {
	stmt, params, err := qb.Delete(OtpTable).Where(sq.Eq{"phone": otp.Phone}).Where(sq.Eq{"type": otp.OtpType}).ToSql()

	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, stmt, params...)
	return err
}

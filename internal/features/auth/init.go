package auth

import (
	"github.com/NuEventTeam/events/internal/config"
	"github.com/NuEventTeam/events/internal/features/sms_provider"
	"github.com/NuEventTeam/events/internal/storage/database"
)

type Auth struct {
	db          *database.Database
	jwt         config.JWT
	smsProvider *sms_provider.SMSProvider
}

func New(db *database.Database, sms *sms_provider.SMSProvider, cfg config.JWT) *Auth {
	return &Auth{
		db:          db,
		jwt:         cfg,
		smsProvider: sms,
	}
}

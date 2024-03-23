package sms_provider

import (
	"context"
	"github.com/NuEventTeam/events/internal/config"
	"github.com/NuEventTeam/events/pkg"
	"log"
	"net/url"
)

type SMSProvider struct {
	pkg.Request
	URL      string
	Login    string
	Password string
	Enabled  bool
}

func New(sms config.SMS) *SMSProvider {

	return &SMSProvider{
		URL:      sms.URL,
		Login:    sms.Login,
		Password: sms.Password,
		Enabled:  sms.Enabled,
	}
}

func (s *SMSProvider) Send(ctx context.Context, phone, msg string) error {
	if !s.Enabled {
		return nil
	}

	link, err := url.Parse(s.URL)
	if err != nil {
		return err
	}

	query := link.Query()

	query.Add("login", s.Login)
	query.Add("psw", s.Password)
	query.Add("phones", phone)
	query.Add("mes", msg)
	query.Add("charset", "utf-8")
	query.Add("translit", "1")

	link.RawQuery = query.Encode()
	log.Println("-----", link.String())
	request := pkg.Request{
		URL:    link.String(),
		Method: "GET",
	}

	_, err = request.Send()
	return err
}

package event

import (
	"fmt"
	"mime/multipart"
	"time"
)

func requireFormFiled(form *multipart.Form, field string) (string, bool) {
	if len(form.Value[field]) == 0 {
		return fmt.Sprintf("request does not contain %s field", field), false
	}
	return "", true
}
func (c *CreateEventRequest) Validate() []string {
	m := []string{}

	if c.Date.Before(time.Now()) {
		m = append(m, "date cannot be set before current date")
	}

	if c.StartsAt.Before(time.Now().Add(time.Hour * 2)) {
		m = append(m, "event start time mast be at least 2 hour before creations")
	}

	if c.EndsAt.Before(time.Time(c.StartsAt)) {
		m = append(m, "improper ending time")
	}
	if len(m) > 0 {
		return m
	}
	return nil
}

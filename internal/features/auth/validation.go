package auth

import (
	"log"
	"regexp"
)

type ValidationError struct {
	Message string
	Field   string
	Tag     string
}

var (
	phoneRegexpKZ = `^77[0-9]{2}[0-9]{3}[0-9]{4}$`
)

func (ve *ValidationError) Error() string {
	return ve.Message
}

func ValidateLength(n int, s string) *ValidationError {
	if len(s) < n {
		return &ValidationError{
			Message: "password length is less that 3",
			Field:   "password",
		}
	}
	return nil
}

func ValidatePhoneNumber(number string) *ValidationError {
	if len(number) != 11 {
		return &ValidationError{
			Message: "Phone number should be equal to 11 digits",
			Field:   "phone",
		}
	}

	done, err := regexp.MatchString(phoneRegexpKZ, number)
	if err != nil {
		log.Println(err)
		return &ValidationError{
			Message: "cannot string match",
			Field:   "phone",
		}
	}
	if !done {
		return &ValidationError{
			Message: "Incorrect phone format for KZ",
			Field:   "phone",
		}
	}
	return nil
}

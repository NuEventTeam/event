package event

import (
	"fmt"
	"mime/multipart"
)

func requireFormFiled(form *multipart.Form, field string) (string, bool) {
	if len(form.Value[field]) == 0 {
		return fmt.Sprintf("request does not contain %s field", field), false
	}
	return "", true
}

package repository

import (
	models "github.com/thteam47/server_management/model"
)

type EmailRepository interface {
	Send(email *models.Email) error
}

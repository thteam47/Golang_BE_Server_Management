package repoimpl

import (
	"fmt"
	"net/smtp"

	models "github.com/thteam47/server_management/model"
	repo "github.com/thteam47/server_management/repository"
)

type EmailRepositoryImpl struct {
}

func NewEmailRepo() repo.EmailRepository {
	return &EmailRepositoryImpl{}
}
func (e *EmailRepositoryImpl) Send(email *models.Email) error {
	from := vi.GetString("from")
	password := vi.GetString("password")
	smtpHost := vi.GetString("smtpHost")
	smtpPort := vi.GetString("smtpPort")
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, email.To, []byte(email.Body))
	if err != nil {
		return fmt.Errorf("Error getting response: %s", err)
	}
	return nil
}

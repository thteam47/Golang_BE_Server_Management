package repository

type OperationRepository interface {
	Connect(username string, password string) error
	Disconnect(idServer string) error
	Export(check bool, limitpage int64, numberpage int64) string
	SendMail()
}
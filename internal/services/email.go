package services

type EmailModel interface {
	Insert(email string) error
	GetAll() ([]string, error)
}

type EmailServiceImpl struct {
	emailModel EmailModel
}

func NewEmailService(emailModel EmailModel) *EmailServiceImpl {
	return &EmailServiceImpl{emailModel: emailModel}
}

func (es *EmailServiceImpl) Create(email string) error {
	return es.emailModel.Insert(email)
}

func (es *EmailServiceImpl) GetAll() ([]string, error) {
	return es.emailModel.GetAll()
}

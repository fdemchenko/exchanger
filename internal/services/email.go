package services

type EmailRepository interface {
	Insert(email string) error
	GetAll() ([]string, error)
}

type EmailServiceImpl struct {
	emailRepository EmailRepository
}

func NewEmailService(emailRepository EmailRepository) *EmailServiceImpl {
	return &EmailServiceImpl{emailRepository: emailRepository}
}

func (es *EmailServiceImpl) Create(email string) error {
	return es.emailRepository.Insert(email)
}

func (es *EmailServiceImpl) GetAll() ([]string, error) {
	return es.emailRepository.GetAll()
}

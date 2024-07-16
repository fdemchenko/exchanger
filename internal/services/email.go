package services

import "strings"

type EmailRepository interface {
	Insert(email string) error
	GetAll() ([]string, error)
	DeleteByEmail(email string) error
	DeleteByID(id int) error
}

type emailServiceImpl struct {
	emailRepository EmailRepository
}

func NewEmailService(emailRepository EmailRepository) *emailServiceImpl {
	return &emailServiceImpl{emailRepository: emailRepository}
}

func (es *emailServiceImpl) Create(email string) error {
	// email is case insensitive
	email = strings.ToLower(email)
	return es.emailRepository.Insert(email)
}

func (es *emailServiceImpl) GetAll() ([]string, error) {
	return es.emailRepository.GetAll()
}

func (es *emailServiceImpl) DeleteByEmail(email string) error {
	// email is case insensitive
	email = strings.ToLower(email)
	return es.emailRepository.DeleteByEmail(email)
}

func (es *emailServiceImpl) DeleteByID(id int) error {
	return es.emailRepository.DeleteByID(id)
}

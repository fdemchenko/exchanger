package models

import "database/sql"

type EmailModel struct {
	DB *sql.DB
}

func (em *EmailModel) Insert(email string) error {
	stmt := `INSERT INTO emails (email) VALUES ($1)`

	_, err := em.DB.Exec(stmt, email)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"emails_email_key\"" {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (em *EmailModel) GetAll() ([]string, error) {
	return nil, nil
}

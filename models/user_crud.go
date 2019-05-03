package models

import (
	"auth_service_template/logger"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gocql/gocql"
)

var (
	ErrUserNotFound      = errors.New("User not found")
	ErrUserAlreadyExists = errors.New("User already exists")
	ErrDbTypeError       = errors.New("DB type error")
)

func (db *DB) ChangeCheme() {
	fmt.Println(db.conn.(*gocql.Session).Query(`ALTER TABLE users (
		ALTER id TYPE uuid;`).Exec())
}

func (db *DB) AddUser(user *User) error {
	switch c := db.conn.(type) {
	case *sql.DB:
		// some logic
		return nil
	case *gocql.Session:
		var query string
		if len(user.Emails) > 0 {
			query = fmt.Sprintf(`SELECT id FROM users WHERE emails CONTAINS '%s' AND login='%s' ALLOW FILTERING`,
				user.Emails[0], user.Login)
		} else {
			query = fmt.Sprintf(`SELECT id FROM users WHERE login='%s' ALLOW FILTERING`, user.Login)
		}
		var id string
		if err := c.Query(
			query).Scan(
			&id); err != nil {
			fmt.Println(err.Error())
			if err != gocql.ErrNotFound {
				db.log.Println(logger.NewError(
					"cassandra",
					fmt.Sprintf("SELECT User error: [%s]", err.Error()),
					logger.ERROR,
				))
				return err
			} else {
				err := c.Query(`INSERT INTO users (id, first_name, last_name, login, phones, age, password, emails) 
		VALUES (uuid(), ?, ?, ?, ?, ?, ?, ?);`,
					(*user).FirstName, (*user).LastName, (*user).Login,
					(*user).Phones, (*user).Age, (*user).Password, (*user).Emails).Exec()
				if err == nil {
					return nil
				} else {
					db.log.Println(logger.NewError(
						"cassandra",
						fmt.Sprintf("INSERT User error: [%s]", err.Error()),
						logger.ERROR,
					))
					return err
				}
			}
		} else {
			return ErrUserAlreadyExists
		}
	}
	return ErrDbTypeError
}

func (db *DB) FindUserByLoginOrEmail(loginField *string) (*User, error) {
	switch c := db.conn.(type) {
	case *sql.DB:
		// some logic
		return nil, nil
	case *gocql.Session:
		user := &User{}
		m := map[string]interface{}{}

		if err := c.Query(
			fmt.Sprintf(`SELECT * FROM users WHERE emails CONTAINS '%s' ALLOW FILTERING`, *loginField)).
			MapScan(m); err != nil {
			if err != gocql.ErrNotFound {
				db.log.Println(logger.NewError(
					"cassandra",
					fmt.Sprintf("Find User error: [%s]", err.Error()),
					logger.ERROR,
				))
				return nil, err
			} else {
				if err = c.Query(
					fmt.Sprintf(`SELECT * FROM users WHERE login='%s' ALLOW FILTERING`, *loginField)).
					MapScan(m); err != nil {
					if err != gocql.ErrNotFound {
						db.log.Println(logger.NewError(
							"cassandra",
							fmt.Sprintf("Find User error: [%s]", err.Error()),
							logger.ERROR,
						))
						return nil, err
					} else {
						return nil, ErrUserNotFound
					}
				} else {
					//	fmt.Println(reflect.TypeOf(m["id"]).String())
					user.ID = m["id"].(gocql.UUID).String()
					user.FirstName = m["first_name"].(string)
					user.LastName = m["last_name"].(string)
					user.Login = m["login"].(string)
					user.Phones = m["phones"].([]string)
					user.Age = int32(m["age"].(int))
					user.Password = m["password"].(string)
					user.Emails = m["emails"].([]string)
					return user, nil
				}
				return nil, ErrUserNotFound
			}
		} else {
			user.ID = m["id"].(gocql.UUID).String()
			user.FirstName = m["first_name"].(string)
			user.LastName = m["last_name"].(string)
			user.Login = m["login"].(string)
			user.Phones = m["phones"].([]string)
			user.Age = int32(m["age"].(int))
			user.Password = m["password"].(string)
			user.Emails = m["emails"].([]string)
			return user, nil
		}
	}
	return nil, ErrDbTypeError
}

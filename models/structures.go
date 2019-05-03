package models

type User struct {
	ID        string   `json:"id", db:"id"`
	FirstName string   `json:"first_name", db:"first_name"`
	LastName  string   `json:"last_name", db:"last_name"`
	Login     string   `json:"login", db:"login"`
	Phones    []string `json:"phones", db:"phones"`
	Age       int32    `json:"age", db:"age"`
	Password  string   `json:"password", db:"password"`
	Emails    []string `json:"emails", db:"emails"`
}

type Credentials struct {
	Password string `json:"password", db:"password"`
	Login    string `json:"login", db:"login"`
}

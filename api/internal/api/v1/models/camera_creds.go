package models

type CameraCreds struct {
	UUID     string `json:"uuid" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `db:"password""`
}

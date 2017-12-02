package dao

import ()

func GetUser(username string) *User {
	var dbUsr User
	Db.First(&dbUsr, "user_id=?", username)
	return &dbUsr
}

package dao

func GetUser(username string) *User {
	var dbUsr User
	Db.First(&dbUsr, "user_id=?", username)
	debug(dbUsr)
	return &dbUsr
}

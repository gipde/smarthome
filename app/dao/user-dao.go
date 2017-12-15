package dao

import ()
import "github.com/revel/revel"

func GetUser(username string) *User {
	var dbUsr User
	Db.First(&dbUsr, "user_id=?", username)
	if dbUsr.UserID != username {
		return nil
	}
	revel.AppLog.Infof("we found User: %+v", dbUsr)
	return &dbUsr
}

func GetUserWithID(uid int) *User {
	var dbUsr User
	Db.First(&dbUsr, "id=?", uid)
	return &dbUsr
}

func SaveUser(user *User) {
	Db.Save(&user)
}

func GetAllUsers() *[]User {
	var users []User
	Db.Find(&users)
	return &users
}

func DeleteUser(user *User) {
	Db.Delete(&user)
}

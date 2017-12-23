package dao

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

// GetUser returns a user
func GetUser(username string) *User {
	var dbUsr User
	Db.First(&dbUsr, "user_id=?", username)
	if dbUsr.UserID != username {
		return nil
	}
	return &dbUsr
}

// GetUserWithID returns a User
func GetUserWithID(uid uint) *User {
	var dbUsr User
	Db.First(&dbUsr, "id=?", uid)
	return &dbUsr
}

// SaveUser saves a user
func SaveUser(user *User) {
	Db.Save(&user)
}

// GetAllUsers returns all User
func GetAllUsers() *[]User {
	var users []User
	Db.Find(&users)
	return &users
}

// DeleteUser removes a User
func DeleteUser(user *User) {
	Db.Delete(&user)
}

// CheckCredentials check if a user/pass combination is correct
func CheckCredentials(username, password string) error {
	dbUsr := GetUser(username)
	if username != dbUsr.UserID {
		return errors.New("user false")
	}
	err := bcrypt.CompareHashAndPassword(dbUsr.Password, []byte(password))
	if err != nil {
		return errors.New("password false")
	}
	return nil
}

// GetOrCreateUser gets a user, if the user do not exists, it will be createtd
func GetOrCreateUser(userid, username string) *User {
	dbUsr := GetUser(userid)
	if dbUsr == nil {
		dbUsr = &User{
			UserID: userid,
			Name:   username,
		}
		SaveUser(dbUsr)
	}
	return dbUsr
}

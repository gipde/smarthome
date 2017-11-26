package app

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	// "github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `gorm:"primary_key"`
	Name     string
	Password []byte
}

var Db *gorm.DB

func InitDB() {

	DbInt, err := gorm.Open("sqlite3", "$HOME/tmp/test.db")

	if err != nil {
		panic("failed to connect database")
	}
	// defer func() {
	// 	revel.AppLog.Info("We Close the DB")
	// 	Db.Close()
	// }()

	Db = DbInt.Debug().AutoMigrate(&User{})

	user := User{Name: "Werner Schneider", ID: "werner"}
	user.Password, _ = bcrypt.GenerateFromPassword([]byte("starten"), bcrypt.DefaultCost)
	Db.Create(user)

}

package dao

import (
	// "fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
	UserID   string
	Name     string
	Password []byte
	Devices  []Device
}

type Device struct {
	gorm.Model
	UserID      int
	Name        string
	Description string
	Producer    string
	DeviceType  int
}

var Db *gorm.DB

func InitDB() {
	revel.AppLog.Info("Init DB")

	driver, _ := revel.Config.String("db.driver")
	connection, _ := revel.Config.String("db.connection")
	DbInt, err := gorm.Open(driver, connection)

	if err != nil {
		panic("failed to connect database")
	}

	Db = DbInt.Debug().AutoMigrate(
		&User{},
		&Device{},
	)

	user := User{Name: "Werner Schneider", UserID: "werner"}
	user.Password, _ = bcrypt.GenerateFromPassword([]byte("starten"), bcrypt.DefaultCost)
	Db.Create(&user)
	user.Devices = *GetTestDevices()
	Db.Save(&user)

}

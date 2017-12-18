package dao

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	gorm.Model
	UserID         string
	Name           string
	Password       []byte
	Devices        []Device
	Authorizations []AuthorizeEntry
}

type Device struct {
	gorm.Model
	UserID      uint
	Name        string
	Description string
	Producer    string
	DeviceType  int
	State       string // eg a json fragment
	Connected   bool
}

type AuthorizeEntry struct {
	gorm.Model
	UserID uint
	AppID  string
}

type Token struct {
	gorm.Model
	Expiry  time.Time
	Code    string
	PayLoad []byte
}

var Db *gorm.DB

func InitDB() {
	revel.AppLog.Info("Init DB")

	driver, _ := revel.Config.String("db.driver")
	connection, _ := revel.Config.String("db.connection")
	dbInt, err := gorm.Open(driver, connection)

	if err != nil {
		panic("failed to connect database")
	}

	dbInt2 := dbInt
	if r, _ := revel.Config.Bool("db.debug"); r {
		dbInt2 = dbInt.Debug()
	}
	Db = dbInt2.AutoMigrate(
		&User{},
		&Device{},
		&AuthorizeEntry{},
		&Token{},
	)

	adminuser, _ := revel.Config.String("user.admin")
	if admin := GetUser(adminuser); admin == nil {
		user := User{Name: adminuser, UserID: adminuser}
		user.Password, _ = bcrypt.GenerateFromPassword([]byte(adminuser), bcrypt.DefaultCost)
		Db.Create(&user)
		user.Devices = *GetTestDevices()
		Db.Save(&user)
	}

}

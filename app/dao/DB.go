package dao

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
)

var Db *gorm.DB

func init() {
	revel.AppLog.Debug("Init")
	revel.OnAppStart(InitDB)
}

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

	// Do Migrations
	Db = dbInt2.AutoMigrate(
		&User{},
		&Device{},
		&AuthorizeEntry{},
		&Token{},
		&Log{},
		&Schedule{},
	)

	// Create Adminuser
	adminuser, _ := revel.Config.String("user.admin")
	if admin := GetUser(adminuser); admin == nil {
		user := User{
			Name:   adminuser,
			UserID: adminuser,
		}
		user.Password, _ = bcrypt.GenerateFromPassword([]byte(adminuser), bcrypt.DefaultCost)
		user.DevicePassword, _ = bcrypt.GenerateFromPassword([]byte(adminuser), bcrypt.DefaultCost)

		Db.Create(&user)

		// save Testdevices with Admin
		user.Devices = *GetTestDevices()
		Db.Save(&user)
	}

}

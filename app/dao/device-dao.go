package dao

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"strconv"
	"strings"
)

func GetAllDevices(user string) *[]Device {
	return getAllDevicesIntern(user, Db)
}
func GetAllDevicesDeep(user string) *[]Device {
	return getAllDevicesIntern(user, Db)
}

func CreateDevice(device *Device) {
	revel.AppLog.Info(fmt.Sprintf("Create Device: %+v", device))
	Db.Create(&device)
}

func DeleteDevice(device *Device) {
	Db.Delete(&device)
}

func FindDeviceByID(user, id string) *Device {
	var device Device
	numericID, _ := strconv.Atoi(strings.TrimPrefix(id, "device-"))

	Db.Where("user_id = ?", user).Find(&device, numericID)

	return &device
}

func getAllDevicesIntern(user string, db *gorm.DB) *[]Device {
	var devices []Device
	db.Where("user_id = ?", user).Find(&devices)
	return &devices
}

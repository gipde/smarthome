package dao

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"strconv"
	"strings"
)

func GetAllDevices(useroid uint) *[]Device {
	return getAllDevicesIntern(useroid, Db)
}
func GetAllDevicesDeep(useroid uint) *[]Device {
	return getAllDevicesIntern(useroid, Db)
}

func CreateDevice(device *Device) {
	revel.AppLog.Info(fmt.Sprintf("Create Device: %+v", device))
	Db.Create(&device)
}

func DeleteDevice(device *Device) {
	Db.Delete(&device)
}

func FindDeviceByID(user uint, id string) *Device {
	var device Device
	numericID, _ := strconv.Atoi(strings.TrimPrefix(id, "device-"))

	Db.Where("user_id = ?", user).Find(&device, numericID)

	return &device
}

func SaveDevice(dev *Device) {
	Db.Save(dev)
}

func getAllDevicesIntern(useroid uint, db *gorm.DB) *[]Device {
	var devices []Device
	db.Where("user_id = ?", useroid).Find(&devices)
	return &devices
}

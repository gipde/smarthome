package dao

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
	"strconv"
	"strings"
)

// GetAllDevices returns a array of all Devices
func GetAllDevices(useroid uint) *[]Device {
	return getAllDevicesIntern(useroid, Db)
}

// CreateDevice ...
func CreateDevice(device *Device) {
	revel.AppLog.Info(fmt.Sprintf("Create Device: %+v", device))
	Db.Create(&device)
}

// DeleteDevice ...
func DeleteDevice(device *Device) {
	Db.Delete(&device)
}

// FindDeviceByID returns device with ID
func FindDeviceByID(user uint, id string) *Device {
	var device Device
	numericID, _ := strconv.Atoi(strings.TrimPrefix(id, "device-"))

	Db.Where("user_id = ?", user).Find(&device, numericID)

	return &device
}

// SaveDevice ...
func SaveDevice(dev *Device) {
	Db.Save(dev)
}

func getAllDevicesIntern(useroid uint, db *gorm.DB) *[]Device {
	var devices []Device
	db.Where("user_id = ?", useroid).Find(&devices)
	return &devices
}

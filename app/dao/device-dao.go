package dao

import (
	"fmt"
	"github.com/revel/revel"
)

func debug(val interface{}) {
	revel.AppLog.Info(fmt.Sprintf("%+v", val))
}

func GetAllDevices() *[]Device {
	var devices []Device
	Db.Preload("AlexaInterfaces").Preload("DisplayCategories").Find(&devices)
	for _, d := range devices {
		revel.AppLog.Info(fmt.Sprintf("%+v", d))
	}
	return &devices
}

func CreateDevice(device *Device) {
	revel.AppLog.Info(fmt.Sprintf("Create Device: %+v", device))
	Db.Create(&device)
}

func DeleteDevice(device *Device) {
	Db.Delete(&device)
}

func FindDeviceByID(id string) *Device {
	var device Device
	Db.Find(&device, id)
	return &device
}

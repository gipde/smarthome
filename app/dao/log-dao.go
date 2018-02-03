package dao

import (
	"time"
)

const (
	logError = iota
	logWarning
	logInfo
	logDebug

	defaultLogLevel = logInfo
)

// PersistLog persist a Log Entry
func PersistLog(device uint, message string) {
	Db.Save(&Log{
		Level:    defaultLogLevel,
		Message:  message,
		DeviceID: device,
		Time:     time.Now(),
	})
}

func GetLogs(device uint) *[]Log {
	var logs []Log
	Db.Where("device_id = ?", device).Find(&logs)
	return &logs
}

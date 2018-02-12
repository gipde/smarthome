package dao

import (
	"github.com/revel/revel"
	"schneidernet/smarthome/app"

	"time"
)

var Days = [...]string{
	"Sonntag",
	"Montag",
	"Dienstag",
	"Mittwoch",
	"Donnerstag",
	"Freitag",
	"Samstag",
}

func GetSchedules(deviceID uint) []Schedule {
	var result []Schedule
	Db.Find(&result, Schedule{DeviceID: deviceID})
	return result
}

func GetSchedule(id uint) *Schedule {
	var result Schedule
	Db.First(&result, id)
	return &result
}

func SaveSchedule(sched *Schedule) {
	Db.Save(sched)
}

func GetSchedulesForTime(hour, minute int) *[]Schedule {
	var schedules []Schedule
	Db.Where("next_run < ?", time.Now()).Order("next_run asc").Find(&schedules)
	return &schedules
}

func CreateCountDown(nextRun time.Time, state string, device uint) *Schedule {
	return createSheduleIntern(nextRun, state, device, true)
}

func createSheduleIntern(nextRun time.Time, state string, device uint, onetime bool) *Schedule {
	return &Schedule{
		NextRun:  nextRun,
		DeviceID: device,
		State:    state,
		OneTime:  onetime,
	}
}

func CreateSchedule(weekday, exectime, state string, device uint, onetime bool) *Schedule {
	etime, err := time.Parse("15:04", exectime)
	if err != nil {
		revel.AppLog.Error("unknown Time", "time", exectime)
		return nil
	}

	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), etime.Hour(), etime.Minute(), 0, 0, time.Local)

	// find next weeday matching scheduled weekday
	idx := app.SliceIndex(len(Days), func(i int) bool { return Days[i] == weekday })
	if idx == -1 {
		revel.AppLog.Error("unknown Weekday", "weekday", weekday)
		return nil
	}
	expectedWeekDay := time.Weekday(idx)
	for {
		if nextRun.Weekday() == expectedWeekDay {
			break
		}
		nextRun = nextRun.AddDate(0, 0, 1)
	}

	// if nextRun is in Past
	if nextRun.Unix() < now.Unix() {
		nextRun = nextRun.AddDate(0, 0, 7)
	}

	return createSheduleIntern(nextRun, state, device, onetime)
}

func DeleteSchedule(sched *Schedule) {
	Db.Delete(sched)
}

func validateInput() {

}

func (s *Schedule) Validate(v *revel.Validation) {
	v.Length(s, 1)
}

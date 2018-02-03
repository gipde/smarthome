package tests

import (
	"github.com/revel/revel"
	"github.com/revel/revel/testing"
	"schneidernet/smarthome/app/dao"
	"time"
)

type ScheduleTest struct {
	testing.TestSuite
}

func (t *ScheduleTest) TestSchedule() {
	current := time.Now()
	schedTime := current.Add(time.Minute * -1).Format("15:04")
	sched := dao.CreateSchedule(dao.Days[current.Weekday()], schedTime, "ON", 1, true)
	revel.AppLog.Info("Times: ", "current", current, "nextRun", sched.NextRun)
	t.Assert(sched.State == "ON")
	t.Assert(sched.DeviceID == 1)
	t.Assert(sched.LastRun == time.Time{})
	t.Assert(sched.NextRun.Weekday() == current.Weekday())
	t.Assert(sched.NextRun.Hour() == current.Hour())
	t.Assert(sched.NextRun.Minute() == current.Add(time.Minute*-1).Minute())
	t.Assert(sched.NextRun.Year() == current.AddDate(0, 0, 7).Year())
	t.Assert(sched.NextRun.Month() == current.AddDate(0, 0, 7).Month())
	t.Assert(sched.NextRun.Day() == current.AddDate(0, 0, 7).Day())

	schedTime = current.Add(time.Minute * 1).Format("15:04")
	sched = dao.CreateSchedule(dao.Days[current.Weekday()], schedTime, "OFF", 1, true)
	t.Assert(sched.State == "OFF")
	t.Assert(sched.DeviceID == 1)
	t.Assert(sched.LastRun == time.Time{})
	t.Assert(sched.NextRun.Weekday() == current.Weekday())
	t.Assert(sched.NextRun.Hour() == current.Hour())
	t.Assert(sched.NextRun.Minute() == current.Add(time.Minute*1).Minute())
	t.Assert(sched.NextRun.Year() == current.AddDate(0, 0, 0).Year())
	t.Assert(sched.NextRun.Month() == current.AddDate(0, 0, 0).Month())
	t.Assert(sched.NextRun.Day() == current.AddDate(0, 0, 0).Day())

	sched = dao.CreateSchedule("hans", schedTime, "OFF", 1, true)
	t.Assert(sched == nil)
	sched = dao.CreateSchedule("Samstag", "wurst", "OFF", 1, true)
	t.Assert(sched == nil)
}

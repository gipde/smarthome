package controllers

import (
	"github.com/revel/revel"
	"schneidernet/smarthome/app/dao"
	"time"
)

func init() {
	revel.AppLog.Debug("Init")
	revel.OnAppStart(executor)
}

func executor() {
	interval := time.Second * 59
	start := time.Now().Add(interval)
	go func() {
		for {

			// wait a Minute
			for {
				if start.Unix() < time.Now().Unix() {
					start = start.Add(interval)
					break
				}
				time.Sleep(time.Second)
			}

			revel.AppLog.Info("we look for schedules...")
			now := time.Now()
			schedules := dao.GetSchedulesForTime(now.Hour(), now.Minute())
			for _, sched := range *schedules {
				SetState(sched.DeviceID, sched.State)
				if !sched.OneTime {
					sched.LastRun = now
					sched.NextRun = now.AddDate(0, 0, 7)
					dao.SaveSchedule(&sched)
				} else {
					dao.DeleteSchedule(&sched)
				}
			}
		}
	}()
}

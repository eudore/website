package schedule

import (
	"github.com/robfig/cron"
)

type ScheduleController struct {
	Cron *cron.Cron
}

func (schedule *ScheduleController) PutNew() {}
func (schedule *ScheduleController) PostById() {
	id := schedule.GetParamInt("id")
	schedule.Cron.Remove(id)
}

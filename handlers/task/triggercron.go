package task

import (
	"github.com/robfig/cron"
)

type (
	CronTrigger struct {
		Cron     *cron.Cron
		Schedule chan *Event
		EventIds map[int]cron.EntryID
	}
)

func NewCronTrigger(sc chan *Event) *CronTrigger {
	crontab := cron.New(cron.WithParser(cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)))
	go crontab.Run()
	return &CronTrigger{
		Cron:     crontab,
		Schedule: sc,
		EventIds: make(map[int]cron.EntryID),
	}
}

func (tg *CronTrigger) AddEntry(i map[string]interface{}) error {
	e := NewEvent(i)
	id, err := tg.Cron.AddFunc(e.Params["cron"].(string), func() {
		tg.Schedule <- e
	})
	if err != nil {
		return err
	}
	tg.EventIds[e.Id] = id
	return nil
}

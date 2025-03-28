package cronjob

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

type CronJobScheduler struct {
	cron *cron.Cron
}

var cronJobInstance *CronJobScheduler

func GetCronJobScheduler() *CronJobScheduler {
	if cronJobInstance != nil {
		return cronJobInstance
	}
	return newCronJobScheduler()
}

func newCronJobScheduler() *CronJobScheduler {
	c := cron.New()
	c.Start()
	return &CronJobScheduler{cron: c}
}

func (t *CronJobScheduler) Schedule(schedule string, task func()) (int, error) {
	id, err := t.cron.AddFunc(schedule, task)
	if err != nil {
		return 0, fmt.Errorf("failed to schedule job: %s %v", schedule, err)
	}
	return int(id), nil
}

func (t *CronJobScheduler) Stop(taskId int) {
	t.cron.Remove(cron.EntryID(taskId))
}

package cronjob_test

import (
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ripple-mq/ripple-server/internal/cronjob"
)

func TestGetCronJobScheduler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "check smooth start",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cronjob.GetCronJobScheduler(); got == nil {
				t.Errorf("GetCronJobScheduler() = %v", got)
			}
		})
	}
}

func TestCronJobScheduler_Schedule(t *testing.T) {
	type args struct {
		schedule string
		task     func()
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "run at every second",
			args: args{schedule: "@every 1s", task: func() {
				log.Infof("Running test job")
			}},
			wantErr: false,
		},
		{
			name: "invalid schedule string",
			args: args{schedule: "100seconds", task: func() {
				log.Infof("Running test job")
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := cronjob.GetCronJobScheduler()
			_, err := tr.Schedule(tt.args.schedule, tt.args.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("CronJobScheduler.Schedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			time.Sleep(3 * time.Second)
		})
	}
}

func TestCronJobScheduler_Stop(t *testing.T) {
	type args struct {
		schedule string
		task     func()
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid task ID",
			args: args{schedule: "@every 1s", task: func() {
				log.Infof("Running test job: Stop")
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := cronjob.GetCronJobScheduler()
			id, err := tr.Schedule(tt.args.schedule, tt.args.task)
			if err != nil {
				t.Errorf("CronJobScheduler.Stop() schedule failed, error = %v, wantErr %v", err, tt.wantErr)
			}
			tr.Stop(id)
		})
	}
}

package api_models

import "time"

type AutomaticJob struct {
	JobName string `gorm:"column:job_name;primary_key"`
	Action string `gorm:"column:action"`
	Frequency time.Duration `gorm:"column:frequency"`
	LastRun time.Time `gorm:"column:last_run"`
	NextRun time.Time `gorm:"column:next_run"`
	Parameters string `gorm:"column:parameters"`
	Count uint64 `gorm:"column:count"`

	Enabled bool `gorm:"column:enabled"`
}

func (AutomaticJob) ColumnName() string {
	return "automaticJobs"
}
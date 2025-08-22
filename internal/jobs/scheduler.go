package jobs

import "time"

type Scheduler interface {
	Schedule(at time.Time, key string, payload any) error
}

type NoopScheduler struct{}

func (NoopScheduler) Schedule(at time.Time, key string, payload any) error {
	return nil
}

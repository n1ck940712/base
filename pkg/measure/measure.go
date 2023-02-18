package measure

import "time"

type execution interface {
	Done() time.Duration
}

type exec struct {
	start time.Time
}

func NewExecution() execution {
	return &exec{start: time.Now()}
}

func (e *exec) Done() time.Duration {
	return time.Since(e.start)
}

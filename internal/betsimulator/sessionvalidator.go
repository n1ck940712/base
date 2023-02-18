package betsimulator

import (
	"time"
)

type checkingType string

const (
	checkingTypeDay    checkingType = "per_day"
	checkingTypeHour   checkingType = "per_hour"
	checkingTypeMinute checkingType = "per_minute"
	checkingTypeSecond checkingType = "per_second"
)

const (
	oneDay    time.Duration = oneHour * 24
	oneHour   time.Duration = time.Duration(1) * time.Hour
	oneMinute time.Duration = time.Duration(1) * time.Minute
	oneSecond time.Duration = time.Duration(1) * time.Second
)

type sessionValidator struct {
	cType             checkingType
	isRunning         bool
	validatorCallback *func(time.Time)
}

func newSessionValidator(cType checkingType) *sessionValidator {
	return &sessionValidator{cType, false, nil}
}

func (sv *sessionValidator) run() {
	if sv.validatorCallback != nil {
		(*sv.validatorCallback)(timeNow())
	}
	go sv.validatorRun()
}

func (sv *sessionValidator) validatorRun() {
	if sv.isRunning {
		return
	}
	sv.isRunning = true
	for sv.isRunning {
		tDuration := sv.remainingTime()

		time.Sleep(tDuration)
		if sv.isRunning && sv.validatorCallback != nil {
			(*sv.validatorCallback)(timeNow())
		}
	}
	sv.isRunning = false
}

func (sv *sessionValidator) stop() {
	if !sv.isRunning {
		return
	}
	sv.isRunning = false
}

func (sv *sessionValidator) setValidatorCallback(callback func(time.Time)) {
	sv.validatorCallback = &callback
}

func (sv *sessionValidator) remainingTime() time.Duration {
	now := timeNow()
	nextHour := now.Truncate(sv.timeDuration()).Add(sv.timeDuration())

	return nextHour.Sub(now)
}

func (sv *sessionValidator) timeDuration() time.Duration {
	switch sv.cType {
	case checkingTypeDay:
		return oneDay
	case checkingTypeHour:
		return oneHour
	case checkingTypeMinute:
		return oneMinute
	case checkingTypeSecond:
		return oneSecond
	default:
		return oneHour
	}
}

func timeNow() time.Time {
	return time.Now() //add UTC if needed
}

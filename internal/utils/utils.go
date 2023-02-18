package utils

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
)

func IfElse[T any](condition bool, ifVal T, elseVal T) T {
	if condition {
		return ifVal
	}
	return elseVal
}

func Ptr[T any](v T) *T {
	return &v
}

func TimeNow() time.Time {
	return time.Now()
}

func ElapsedTime(prevTime time.Time) int64 {
	return TimeNow().UnixMilli() - prevTime.UnixMilli()
}

func TimeToUnixTS(time time.Time) float64 {
	return float64(time.UnixMicro()) / 1_000_000
}

func GenerateUnixTS() float64 {
	return TimeToUnixTS(time.Now())
}

func PrettyJSON(a any) string {
	if settings.GetEnvironment().String() == "local" {
		jStr, _ := json.MarshalIndent(a, "", "    ")

		return string(jStr)
	} else {
		return JSON(a)
	}
}

func JSON(a any) string {
	jStr, _ := json.Marshal(a)

	return string(jStr)
}

func CancelContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func TerminateContext() (context.Context, context.CancelFunc) {
	ctx, cancel := CancelContext()
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func(sigTermChan <-chan os.Signal) {
		<-sigTermChan
		cancel()
	}(signalChan)
	return ctx, cancel
}

func Sleep(duration time.Duration, ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-time.After(duration):
	}
}

func PerformAfter(duration time.Duration, toPerform func()) {
	go func(toPerform func()) {
		ctx, _ := TerminateContext()

		select {
		case <-ctx.Done():
			//terminated
		case <-time.After(duration):
			toPerform()
		}
	}(toPerform)
}

// check if a value has a field on list of fields
func ValueHasAField(value any, fields []string) bool {
	rValue := reflect.ValueOf(value).Elem()

	for _, field := range fields {
		if rValue.FieldByName(field).Kind() != reflect.Invalid {
			return true
		}
	}
	return false
}

//new implementation ends

func InTimeSpan(start time.Time, end time.Time, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}

func ContainsAny(value string, subs ...string) (bool, bool, int) {
	nMatches := 0
	isDone := true

	for _, sub := range subs {
		if strings.Contains(value, sub) {
			nMatches += 1
		} else {
			isDone = false
		}
	}

	return isDone, nMatches > 0, nMatches
}

func Contains(list []string, value string) bool {
	for _, a := range list {
		if a == value {
			return true
		}
	}
	return false
}

func InArray(val interface{}, array interface{}) (exists bool) {
	values := reflect.ValueOf(array)

	if reflect.TypeOf(array).Kind() == reflect.Slice || values.Len() > 0 {
		for i := 0; i < values.Len(); i++ {
			if reflect.DeepEqual(val, values.Index(i).Interface()) {
				return true
			}
		}
	}

	return false
}

func CalculateMaxPayout(betAmount float64, euroOdds float64) float64 {
	return betAmount * euroOdds
}

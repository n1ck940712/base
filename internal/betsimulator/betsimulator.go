package betsimulator

import (
	"context"
	"math"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
)

type BetSimulatorPhase string
type BettorsPhase string

const (
	SetBettors           BetSimulatorPhase = "SetBettors"
	StartSessionValidate BetSimulatorPhase = "StartSessionValidate"
	StartBetting         BetSimulatorPhase = "StartBetting"
	StartResulting       BetSimulatorPhase = "StartResulting"
	ActiveBettors        BettorsPhase      = "active"
	WinningBettors       BettorsPhase      = "winning"
)

type BetSimulator interface {
	SetBettors(bettors []BetSimulatorData)
	Bettors() []BetSimulatorData
	SessionBettors() []BetSimulatorData
	StartBetting()
	ActiveBettors() []BetSimulatorData
	SetActiveBettorsCallback(callback func(data BetSimulatorData))
	StartResulting()
	WinningBettors() []BetSimulatorData
	SetWinningBettorsCallback(callback func(data BetSimulatorData))
	SetPhaseCallback(callback func(phase BetSimulatorPhase))
}

type BetSimulatorData interface {
	GetID() string                    //identifier for storing session minutes
	GetName() string                  //name
	GetIsEnabled() bool               //is enabled for betting
	GetSessionProbablility() float64  //probability to start betting per hour
	GetSessionMinutes() time.Duration //session duration in minutes
	GetDaysActive() string            //active days (1 = monday , 2 = tuesday , etc)
	GetMinActiveHMS() string          //"time_active_start":"00:00:01", min hours minutes seconds bet range
	GetMaxActiveHMS() string          //“time_active_end":"23:00:00", max hours minutes seconds bet range
	GetBetProbability() float64       //probality to bet per event max 100
	GetBetWinProbability() float64    //probality to win a bet per event max 100
	GetMinBettingMS() int64           //milliseconds min start bet - 0 to start instant
	GetMaxBettingMS() int64           //milliseconds max start bet - 0 to start instant
	GetMinResultingMS() int64         //milliseconds min start resulting - 0 to start instant
	GetMaxResultingMS() int64         //milliseconds max start resulting - 0 to start instant
}

type betSimulator struct {
	sessValidator          *sessionValidator
	activeSessions         map[string]time.Time //store session using key GetID() and value expiration time GetSessionMinutes() + timeNow()
	bettors                []BetSimulatorData
	activeBettors          []BetSimulatorData
	activeBettorsCallback  *func(data BetSimulatorData)
	winningBettors         []BetSimulatorData
	winningBettorsCallback *func(data BetSimulatorData)
	phaseCallback          *func(phase BetSimulatorPhase)
	sleepCancelCtx         *context.CancelFunc
}

func NewBetSimulator() BetSimulator {
	sessionValidator := newSessionValidator(checkingTypeHour)
	betSimulator := &betSimulator{
		sessionValidator,
		map[string]time.Time{},
		[]BetSimulatorData{},
		[]BetSimulatorData{},
		nil,
		[]BetSimulatorData{},
		nil,
		nil,
		nil,
	}

	sessionValidator.setValidatorCallback(betSimulator.validateSession)
	sessionValidator.run()
	return betSimulator
}

func (bs *betSimulator) validateSession(vTime time.Time) {
	if bs.phaseCallback != nil {
		(*bs.phaseCallback)(StartSessionValidate)
	}
	bs.removeExpiredSessions()
	for _, bsd := range bs.bettors {
		if time, ok := bs.activeSessions[bsd.GetID()]; ok {
			logger.Info("REMAINING time: ", time.Sub(vTime), " for id: ", bsd.GetID())
		} else if bsd.GetSessionProbablility() >= generateProbability(100) {
			bs.activeSessions[bsd.GetID()] = vTime.Add(bsd.GetSessionMinutes())
		}
	}
}

func (bs *betSimulator) SetBettors(bettors []BetSimulatorData) {
	if bs.phaseCallback != nil {
		(*bs.phaseCallback)(SetBettors)
	}
	prevBettors := bs.bettors
	bs.bettors = bettors
	if hasChangesOnBettorIDs(prevBettors, bs.bettors) { //rerun if has changes
		bs.sessValidator.run()
	}
}

func (bs *betSimulator) Bettors() []BetSimulatorData {
	return bs.bettors
}

func (bs *betSimulator) SessionBettors() []BetSimulatorData {
	sessionBettors := []BetSimulatorData{}

	for _, bsd := range bs.bettors {
		if _, ok := bs.activeSessions[bsd.GetID()]; ok {
			sessionBettors = append(sessionBettors, bsd)
		}
	}
	return sessionBettors
}

func (bs *betSimulator) StartBetting() {
	bs.removeExpiredSessions()
	if bs.phaseCallback != nil {
		(*bs.phaseCallback)(StartBetting)
	}
	if bs.sleepCancelCtx != nil { //cancel sleeps
		(*bs.sleepCancelCtx)()
	}
	bs.activeBettors = []BetSimulatorData{}
	ctx, cancel := context.WithCancel(context.Background())
	bs.sleepCancelCtx = &cancel
	tNow := timeNow()

	for _, bsd := range bs.SessionBettors() {
		if bsd.GetIsEnabled() &&
			tNow.After(hmsToRelativeTime(tNow, bsd.GetMinActiveHMS())) && //allow betting if in range of min and max active hms
			tNow.Before(hmsToRelativeTime(tNow, bsd.GetMaxActiveHMS())) &&
			isDaysActive(tNow, bsd.GetDaysActive()) {
			//check probability to bet on event
			if bsd.GetBetProbability() >= generateProbability(100) {
				sleepDuration := generateSleepDuration(bsd.GetMinBettingMS(), bsd.GetMaxBettingMS())

				if sleepDuration > 0 {
					go bs.addToBettors(ActiveBettors, bsd, sleepDuration, ctx)
				} else {
					bs.addToBettors(ActiveBettors, bsd, sleepDuration, ctx)
				}
			}
		}
	}
}

func (bs *betSimulator) ActiveBettors() []BetSimulatorData {
	bs.sessValidator.stop()
	return bs.activeBettors
}

func (bs *betSimulator) SetActiveBettorsCallback(callback func(data BetSimulatorData)) {
	bs.activeBettorsCallback = &callback
}

func (bs *betSimulator) StartResulting() {
	if bs.phaseCallback != nil {
		(*bs.phaseCallback)(StartResulting)
	}
	if bs.sleepCancelCtx != nil { //cancel sleeps
		(*bs.sleepCancelCtx)()
	}
	bs.winningBettors = []BetSimulatorData{}
	ctx, cancel := context.WithCancel(context.Background())

	bs.sleepCancelCtx = &cancel
	for _, bsd := range bs.activeBettors {
		//check probability to win
		if bsd.GetBetWinProbability() >= generateProbability(100) {
			sleepDuration := generateSleepDuration(bsd.GetMinResultingMS(), bsd.GetMaxResultingMS())

			if sleepDuration > 0 {
				go bs.addToBettors(WinningBettors, bsd, sleepDuration, ctx)
			} else {
				bs.addToBettors(WinningBettors, bsd, sleepDuration, ctx)
			}
		}
	}
}

func (bs *betSimulator) WinningBettors() []BetSimulatorData {
	return bs.winningBettors
}

func (bs *betSimulator) SetWinningBettorsCallback(callback func(data BetSimulatorData)) {
	bs.winningBettorsCallback = &callback
}

func (bs *betSimulator) SetPhaseCallback(callback func(phase BetSimulatorPhase)) {
	bs.phaseCallback = &callback
}

func (bs *betSimulator) addToBettors(bPhase BettorsPhase, bsd BetSimulatorData, sleepDuration time.Duration, ctx context.Context) {
	procedure := func() {
		switch bPhase {
		case ActiveBettors:
			bs.activeBettors = append(bs.activeBettors, bsd)
			if bs.activeBettorsCallback != nil { //callback of active bettors
				(*bs.activeBettorsCallback)(bsd)
			}
		case WinningBettors:
			bs.winningBettors = append(bs.winningBettors, bsd)
			if bs.winningBettorsCallback != nil { //callback of winning bettors
				(*bs.winningBettorsCallback)(bsd)
			}
		}
	}

	if sleepDuration > 0 {
		select {
		case <-ctx.Done():
		case <-time.After(sleepDuration * time.Millisecond):
			procedure()
		}
	} else {
		procedure()
	}
}

func (bs *betSimulator) removeExpiredSessions() {
	tActiveSessions := bs.activeSessions
	tNow := timeNow()

	for k, tExpiration := range tActiveSessions {
		if tNow.After(tExpiration) { //if time now is greater than tExpiration
			delete(bs.activeSessions, k) //delete key
		}
	}
}

func isDaysActive(tNow time.Time, days string) bool {
	tDays := daysToDayArr(days)

	for _, tDay := range tDays {
		if tNow.Weekday() == tDay {
			return true
		}
	}
	return false
}

func daysToDayArr(days string) []time.Weekday {
	tDays := []time.Weekday{}

	for _, day := range strings.Split(days, ",") {
		switch day {
		case "0", "7":
			tDays = append(tDays, time.Sunday)
		case "1":
			tDays = append(tDays, time.Monday)
		case "2":
			tDays = append(tDays, time.Tuesday)
		case "3":
			tDays = append(tDays, time.Wednesday)
		case "4":
			tDays = append(tDays, time.Thursday)
		case "5":
			tDays = append(tDays, time.Friday)
		case "6":
			tDays = append(tDays, time.Saturday)
		}
	}
	return tDays
}

func hmsToRelativeTime(tNow time.Time, hms string) time.Time {
	tHMS, err := time.Parse("15:04:05", hms)

	if err != nil {
		logger.Error("hmsToRelativeTime error parsing hms: ", hms, " err: ", err.Error())
	}
	return time.Date(tNow.Year(), tNow.Month(), tNow.Day(), tHMS.Hour(), tHMS.Minute(), tHMS.Second(), tHMS.Nanosecond(), tNow.Location())
}

func generateSleepDuration(min int64, max int64) time.Duration {
	if min == max && max == 0 {
		return time.Duration(0)
	}
	sleepDuration := time.Duration(generateProbability(float64(max)))

	if sleepDuration < time.Duration(min) {
		sleepDuration = time.Duration(min)
	}
	return sleepDuration
}

func generateProbability(max float64) float64 {
	if max <= 0 {
		return 0
	}
	rSource := rand.NewSource(time.Now().UnixNano())
	random := rand.New(rSource)

	return math.Mod(random.Float64()*max, max)
}

func hasChangesOnBettorIDs(bettors1 []BetSimulatorData, bettors2 []BetSimulatorData) bool {
	if len(bettors1) == len(bettors2) {
		for i, bsd1 := range bettors1 {
			if bsd1.GetID() != bettors2[i].GetID() {
				return true
			}
		}
		return false
	}
	return true
}

package betsimulator

import "time"

type FakeMember struct {
	BetProbability     float64 `json:"bet_probability"`
	Name               string  `json:"name"`
	SessionMinute      float64 `json:"session_minute"`
	SessionProbability float64 `json:"session_probability"`
}

func (fm *FakeMember) GetID() string {
	return fm.Name
}

func (fm *FakeMember) GetName() string {
	return fm.Name
}

func (fm *FakeMember) GetSessionProbablility() float64 {
	return 50
}

func (fm *FakeMember) GetSessionMinutes() time.Duration {
	return time.Duration(fm.SessionMinute) * time.Minute
}

func (fm *FakeMember) GetIsEnabled() bool {
	return true
}

func (fm *FakeMember) GetDaysActive() string {
	return "1,2,3,4,5"
}

func (fm *FakeMember) GetMinActiveHMS() string {
	return "02:15:00"
}

func (fm *FakeMember) GetMaxActiveHMS() string {
	return "23:32:00"
}

func (fm *FakeMember) GetBetProbability() float64 {
	return fm.BetProbability
}

func (fm *FakeMember) GetBetWinProbability() float64 {
	return 60
}

func (fm *FakeMember) GetMinBettingMS() int64 {
	return 10
}

func (fm *FakeMember) GetMaxBettingMS() int64 {
	return 7000
}

func (fm *FakeMember) GetMinResultingMS() int64 {
	return 0
}

func (fm *FakeMember) GetMaxResultingMS() int64 {
	return 0
}

package models

import (
	"time"
)

func (FakeMember) TableName() string {
	return "mini_game_fake_member"
}

type FakeMember struct {
	ID                    int64     `json:"id" gorm:"primaryKey;autoIncrement:true"`
	Ctime                 time.Time `json:"-" gorm:"default:current_timestamp; index:idx_mini_game_fake_member_ctime, type:btree"`
	Mtime                 time.Time `json:"-" gorm:"default:current_timestamp; autoUpdateTime:true; index:idx_mini_game_fake_member_mtime, type:btree"`
	UserID                string    `json:"user_id,omitempty" gorm:"index:idx_mini_game_fake_member_user_id, type:btree; uniqueIndex:idx_fake_member_user_table_id"`
	MiniGameTableID       int64     `json:"mini_game_table_id,omitempty" gorm:"index:idx_mini_game_fake_member_mini_game_table_id, type:btree; uniqueIndex:idx_fake_member_user_table_id"`
	IsEnabled             bool      `json:"is_enabled,omitempty"`
	DayActive             string    `json:"day_active,omitempty"`
	BetWinningProbability float64   `json:"bet_winning_probability,omitempty" binding:"min=0,max=100" gorm:"type:numeric(5,2)"`
	SessionProbability    float64   `json:"session_probability,omitempty" binding:"min=0,max=100" gorm:"type:numeric(5,2)"`
	BetProbability        float64   `json:"bet_probability,omitempty" binding:"min=0,max=100" gorm:"type:numeric(5,2)"`
	SessionMinute         int       `json:"session_minute,omitempty"`
	TimeActiveStart       string    `json:"time_active_start,omitempty"`
	TimeActiveEnd         string    `json:"time_active_end,omitempty"`
}

func (fm *FakeMember) GetID() string {
	return fm.UserID
}

func (fm *FakeMember) GetName() string {
	return fm.UserID
}

func (fm *FakeMember) GetSessionProbablility() float64 {
	return fm.SessionProbability
}

func (fm *FakeMember) GetSessionMinutes() time.Duration {
	return time.Duration(fm.SessionMinute) * time.Minute
}

func (fm *FakeMember) GetIsEnabled() bool {
	return fm.IsEnabled
}

func (fm *FakeMember) GetDaysActive() string {
	return fm.DayActive
}

func (fm *FakeMember) GetMinActiveHMS() string {
	return fm.TimeActiveStart
}

func (fm *FakeMember) GetMaxActiveHMS() string {
	return fm.TimeActiveEnd
}

func (fm *FakeMember) GetBetProbability() float64 {
	return fm.BetProbability
}

func (fm *FakeMember) GetBetWinProbability() float64 {
	return fm.BetWinningProbability
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

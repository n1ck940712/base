package response

import (
	"strconv"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
)

type ConfigData struct {
	CurrentEvent    *ConfigEventData `json:"current_event"`
	BetChips        *[]float64       `json:"bet_chips"`
	MinBetAmount    *float64         `json:"min_bet_amount"`
	MaxBetAmount    *float64         `json:"max_bet_amount"`
	MaxPayoutAmount *float64         `json:"max_payout_amount"`
	ShowCharts      *string          `json:"show_charts,omitempty"`
	EnableAutoPlay  *bool            `json:"enable_auto_play,omitempty"`
	IsAnonymous     bool             `json:"is_anonymous,omitempty"`
	Enable          bool             `json:"enable"`
	ResultAnimation bool             `json:"result_animation"`
	EffectsSound    float32          `json:"effects_sound"`
	GameSound       float32          `json:"game_sound"`
	Tour            *bool            `json:"tour,omitempty"`
}

type ConfigEventData struct {
	ID            int64     `json:"id"`
	StartDatetime time.Time `json:"start_datetime"`
}

func (ConfigData) Description() string {
	return "The quick brown"
}

func (cd *ConfigData) SetShowCharts(value string) {
	showCharts := strings.ReplaceAll(value, `"`, "'")

	cd.ShowCharts = &showCharts
}

func (cd *ConfigData) SetResultAnimation(value string) {
	resultAnimation, err := strconv.ParseBool(value)

	if err != nil {
		logger.Error("SetResultAnimation ParseBool error: ", err.Error())
	} else {
		cd.ResultAnimation = resultAnimation
	}
}

func (cd *ConfigData) SetEffectsSound(value string) {
	effectsSound, err := strconv.ParseFloat(value, 32) //32 bit parse float

	if err != nil {
		logger.Error("SetEffectsSound ParseFloat error: ", err.Error())
	} else {
		cd.EffectsSound = float32(effectsSound)
	}
}

func (cd *ConfigData) SetGameSound(value string) {
	gameSound, err := strconv.ParseFloat(value, 32) //32 bit parse float

	if err != nil {
		logger.Error("SetGameSound ParseFloat error: ", err.Error())
	} else {
		cd.GameSound = float32(gameSound)
	}
}

func (cd *ConfigData) SetTour(value string) {
	tour, err := strconv.ParseBool(value)

	if err != nil {
		logger.Error("SetTour ParseBool error: ", err.Error())
	} else {
		cd.Tour = &tour
	}
}

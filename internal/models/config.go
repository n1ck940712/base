package models

import (
	"time"
)

type Config struct {
	Ctime  time.Time `json:"ctime"`
	Mtime  time.Time `json:"mtime"`
	Name   string    `json:"name"`
	Value  string    `json:"value"`
	GameId int       `json:"game_id"`
}
type ConfigPatchRequestBody struct {
	EffectsSound   *float64 `json:"effects_sound,omitempty" binding:"min=0,max=1"`
	GameSound      *float64 `json:"game_sound,omitempty" binding:"min=0,max=1"`
	ResultAimation *bool    `json:"result_animation,omitempty"`
	ShowCharts     *string  `json:"showCharts,omitempty"`
	Tour           *bool    `json:"tour,omitempty"`
}

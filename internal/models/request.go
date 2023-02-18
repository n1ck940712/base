package models

type Request struct {
	EffectsSound    *float64 `json:"effects_sound,omitempty" binding:"min=0,max=1"`
	GameSound       *float64 `json:"game_sound,omitempty" binding:"min=0,max=1"`
	ResultAnimation *bool    `json:"result_animation,omitempty"`
	ShowCharts      *string  `json:"showCharts,omitempty"`
	Tour            *bool    `json:"tour,omitempty"`
}

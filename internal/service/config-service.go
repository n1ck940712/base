package service

import (
	"time"
)

type CurrentEvent struct {
	ID            int64     `json:"id"`
	StartDatetime time.Time `json:"start_datetime"`
}

type Config struct {
	ID              int64        `json:"id"`
	MaxBetAmount    float64      `json:"max_bet_amount"`
	MaxPayoutAmount float64      `json:"max_payout_amount"`
	MinBetAmount    float64      `json:"min_bet_amount"`
	BetChips        []float64    `json:"bet_chips"`
	CurrentEvent    CurrentEvent `json:"current_event"`
}

// type IConfig interface {
// 	GetTableConfig(int64) (Config, map[string]interface{})
// }

func NewIConfig() *Config {
	return &Config{}
}

// {
//     "type": "games.tower.config",
//     "data": {
//         "betChips": [ - meron
//             10,
//             100,
//             500,
//             1000,
//             5000
//         ],
//         "current_event": { - meron
//             "id": 2976961,
//             "start_datetime": "2022-04-21 11:53:15.751264+00:00"
//         },
//         "effects_sound": 0.4,
//         "enable": true,
//         "enable_auto_play": false,
//         "game_sound": 0.36,
//         "is_anonymous": false,
//         "max_bet_amount": "5000.00", - meron
//         "max_payout_amount": "20000.00", - meron
//         "min_bet_amount": "10.00", -meron
//         "result_animation": true
//     }
// }

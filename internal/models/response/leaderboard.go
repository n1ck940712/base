package response

type Leaderboard struct {
	Level string `json:"level,omitempty"`
	Users []User `json:"users,omitempty"`
}

func (Leaderboard) Description() string {
	return "response leaderboard only for json responses"
}

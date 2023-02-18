package response

type State struct {
	Name string  `json:"name,omitempty"`
	End  float64 `json:"end,omitempty"`
	Gts  float64 `json:"gts,omitempty"`
}

func (State) Description() string {
	return "response state only for json responses"
}

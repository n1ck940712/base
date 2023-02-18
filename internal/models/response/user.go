package response

type User struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (User) Description() string {
	return "response user only for json responses"
}

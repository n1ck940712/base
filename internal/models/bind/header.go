package bind

type Header struct {
	Token string `header:"Authorization" form:"auth_token"`
}

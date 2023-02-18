package wsclient

//OnConnectHandler
type OnConnectHandler interface {
	ProcessMessage(msg string)
	WriteMessage(msg string)
}

type onConnectCallback func(handler OnConnectHandler)

//OnConnectHandler - extension
func (wsc *wsclient) ProcessMessage(msg string) {
	response := wsc.GetProcess().ProcessRequest(msg)

	wsc.ws.WriteMessage(response.JSON())
}

func (wsc *wsclient) WriteMessage(msg string) {
	wsc.ws.WriteMessage(msg)
}

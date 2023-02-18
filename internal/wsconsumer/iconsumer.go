package wsconsumer

import "net/http"


type IConsumer interface {
	HandleRequest(w http.ResponseWriter, r *http.Request)
}

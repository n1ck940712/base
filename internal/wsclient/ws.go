package wsclient

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WS interface {
	GetError() error
	GetURLQuery(query string) string
	WriteMessage(msg string) error
	ReadMessage() (string, error)
	Close()
	GetUserRequest() *models.UserRequest
}

type ws struct {
	conn *websocket.Conn
	w    http.ResponseWriter
	r    *http.Request
	err  error
}

func NewWS(w http.ResponseWriter, r *http.Request) WS {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		http.Error(w, "Unable to upgrade connection", http.StatusServiceUnavailable)
		return &ws{err: err}
	}
	return &ws{conn: conn, w: w, r: r}
}

func (ws *ws) GetError() error {
	return ws.err
}

func (ws *ws) GetURLQuery(query string) string {
	if ws.r == nil {
		return ""
	}
	return ws.r.URL.Query().Get(query)
}

func (ws *ws) WriteMessage(msg string) error {
	if ws.conn == nil {
		return errors.New("websocket connetion not initialize")
	}
	if strings.Contains(msg, `"cts_ts"`) {
		msg = fmt.Sprintf(strings.Replace(msg, `"cts_ts"`, "%f", 1), utils.GenerateUnixTS()) //replace %f with current timestamp
	}
	if err := ws.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		return err
	}
	return nil
}

func (ws *ws) ReadMessage() (string, error) {
	if ws.conn == nil {
		return "", errors.New("websocket connetion not initialize")
	}
	mt, raw, err := ws.conn.ReadMessage()

	if err != nil {
		return string(raw), err
	}
	switch mt {
	case websocket.TextMessage:
		return string(raw), nil
	case websocket.CloseMessage:
		return "", errors.New("close message received")
	default:
		return "", nil //ignore other message types
	}
}

func (ws *ws) Close() {
	if ws.conn == nil {
		return
	}
	ws.conn.Close()
}

func (ws *ws) GetUserRequest() *models.UserRequest {
	if ws.r == nil {
		return nil
	}
	return models.NewUserRequest(ws.r)
}

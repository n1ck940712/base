package websocket

import (
	"net/http"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
	wsclient_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/fifashootup"
	wsclient_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/fishprawncrab"
	wsclient_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/lolcouple"
	wsclient_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/loltower"
	"github.com/gorilla/mux"
)

func Run(identifier string, port string) {
	router := mux.NewRouter()
	websocket := newWebsocket(identifier, router)

	websocket.Start()
	go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(identifier)+"websocket", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	defer websocket.Stop()
	http.Handle("/", router)
	if port == "" {
		panic(identifier + " gameloop invalid port: \"" + port + "\"")
	}
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

func newWebsocket(identifier string, router *mux.Router) wsclient.MessageReceiverBroker {
	switch identifier {
	case constants_loltower.Identifier:
		broker := wsclient_loltower.NewReceiverBroker()

		router.HandleFunc("/ws/lol-tower/", func(w http.ResponseWriter, r *http.Request) {
			client := wsclient_loltower.NewWSClient(broker)

			client.Serve(w, r)
		})
		return broker
	case constants_lolcouple.Identifier:
		broker := wsclient_lolcouple.NewReceiverBroker()

		router.HandleFunc("/ws/lol-couple/", func(w http.ResponseWriter, r *http.Request) {
			client := wsclient_lolcouple.NewWSClient(broker)

			client.Serve(w, r)
		})
		return broker
	case constants_fifashootup.Identifier:
		broker := wsclient_fifashootup.NewReceiverBroker()

		router.HandleFunc("/ws/soccer-shootout/", func(w http.ResponseWriter, r *http.Request) {
			client := wsclient_fifashootup.NewWSClient(broker)

			client.Serve(w, r)
		})
		return broker
	case constants_fishprawncrab.Identifier:
		broker := wsclient_fishprawncrab.NewReceiverBroker()

		router.HandleFunc("/ws/fish-prawn-crab/", func(w http.ResponseWriter, r *http.Request) {
			client := wsclient_fishprawncrab.NewWSClient(broker)

			client.Serve(w, r)
		})
		return broker
	default:
		panic("gameloop invalid identifier: \"" + identifier + "\"")
	}
}

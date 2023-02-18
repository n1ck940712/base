package main

import (
	"net/http"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/migrations"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/slack"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient"
	wsclient_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/fifashootup"
	wsclient_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/lolcouple"
	wsclient_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/wsclient/loltower"
	"github.com/gorilla/mux"
)

func main() {
	migrations.NewMigrations().Migrate()

	r := mux.NewRouter()

	loltowerBroker := PrepareLOLTower(r)
	lolcoupleBroker := PrepareLOLCouple(r)
	fifashootupBroker := PrepareFIFAShootup(r)

	defer loltowerBroker.Stop()
	defer lolcoupleBroker.Stop()
	defer fifashootupBroker.Stop()

	http.Handle("/", r)
	http.ListenAndServe("0.0.0.0:"+settings.WS_PORT, nil)
}

func PrepareLOLTower(r *mux.Router) wsclient.MessageReceiverBroker {
	broker := wsclient_loltower.NewReceiverBroker()

	r.HandleFunc("/ws/lol-tower/", func(w http.ResponseWriter, r *http.Request) {
		client := wsclient_loltower.NewWSClient(broker)

		client.Serve(w, r)
	})
	go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(constants_loltower.Identifier)+"websocket", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	broker.Start()
	return broker
}

func PrepareLOLCouple(r *mux.Router) wsclient.MessageReceiverBroker {
	broker := wsclient_lolcouple.NewReceiverBroker()

	r.HandleFunc("/ws/lol-couple/", func(w http.ResponseWriter, r *http.Request) {
		client := wsclient_lolcouple.NewWSClient(broker)

		client.Serve(w, r)
	})
	broker.Start()
	go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(constants_lolcouple.Identifier)+"websocket", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	return broker
}

func PrepareFIFAShootup(r *mux.Router) wsclient.MessageReceiverBroker {
	broker := wsclient_fifashootup.NewReceiverBroker()

	r.HandleFunc("/ws/soccer-shootout/", func(w http.ResponseWriter, r *http.Request) {
		client := wsclient_fifashootup.NewWSClient(broker)

		client.Serve(w, r)
	})
	go slack.SendPayload(slack.NewLootboxNotification(slack.IdentifierToTitle(constants_fifashootup.Identifier)+"websocket", "> *SERVER STARTED*"), slack.LootboxHealthCheck)
	broker.Start()
	return broker
}

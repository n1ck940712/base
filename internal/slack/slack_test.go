package slack

import (
	"fmt"
	"os"
	"testing"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
)

func TestSendMessage(t *testing.T) {
	// if err := SendMessage(":bangbang: *WEB-SOCKET MESSAGES* :bangbang:", "test", LOLTower); err != nil {
	// 	t.Fatalf(err.Error())
	// }

	SendPayload(NewLootboxNotification("lolcouple", "BONUS ROUND ON NEXT EVENT!!!"), 10)
}

func TestSendPayload(t *testing.T) {
	SetEnable(true)
	payload := NewPayload().
		AddSection(func(section Section) Section {
			section.SetText("duterte :punch::skin-tone-3: ", true)
			return section
		}).
		AddHeader("Header!").
		AddImage("Please enjoy this photo of a kitten", "http://placekitten.com/500/500", "cute").
		AddSection(func(section Section) Section {
			section.SetMarkdown("> ```testing 12e asaklf klashflng 12e asaklf klashfla lfhaklshfklahflka fhlkah flkahflk ahfklafklah klf fasfkafsjalfjlasfa dasdasdaa lfhaklshfklahflka fhlkah flkahflk ahfklafklah klf fasfkafsjalfjlasfa dasdasda```")
			section.SetAccessoryButton("push", "push1", "primary")
			section.AddFieldMarkdown("> testing1 ")
			section.AddFieldMarkdown("> testing2 ")
			return section
		}).
		AddActions(func(actions Actions) Actions {
			actions.AddButton("approve", "approve_1", "primary")
			actions.AddButton("deny", "deny_1", "danger")
			actions.AddButton("none", "none_1", "")
			return actions
		}).
		AddSection(func(section Section) Section {
			section.SetText("> testing 1231231231231231231231231231231231231231231231231231231231231234123123123", false)
			section.SetAccessoryImage("https://s3-media3.fl.yelpcdn.com/bphoto/c7ed05m9lC2EmA3Aruue7A/o.jpg", "testing image")
			section.AddFieldMarkdown("> testing1 ")
			section.AddFieldMarkdown("> testing2 ")
			return section
		})
	if err := SendPayload(payload.AsNotification(":bangbang: *WEB-SOCKET MESSAGES* :bangbang:", "#FF0000"), 10); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestSendTemplate(t *testing.T) {
	SetEnable(true)
	os.Setenv("ENVIRONMENT", "local")
	payload := NewLootboxNotification(
		"gameloop",
		"> *SERVER STARTED*")
	if err := SendPayload(payload, 10); err != nil {
		t.Fatalf(err.Error())
	}
	os.Setenv("ENVIRONMENT", "dev")
	payloadDev := NewLootboxNotification(
		"websocket",
		"> *SERVER STARTED*")
	if err := SendPayload(payloadDev, 10); err != nil {
		t.Fatalf(err.Error())
	}
	os.Setenv("ENVIRONMENT", "live")
	payloadLive := NewLootboxNotification(
		"api",
		"> *SERVER STARTED*")
	if err := SendPayload(payloadLive, 10); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestHTTP500(t *testing.T) {
	SetEnable(true)
	if err := api.NewAPI("https://httpstat.us/500").Get(nil); err != nil {
		println("StatusCode : ", err.GetResponseBody())
		println("StatusCode : ", err.GetResponse().StatusCode)
		payload := NewLootboxNotification(
			"api authenticate",
			fmt.Sprint("*Status Code:* \n> *", err.GetResponse().StatusCode, "*\n*Response Body:* \n> *", err.GetResponseBody(), "*"),
		)

		if err := SendPayload(payload, 10); err != nil {
			t.Fatalf(err.Error())
		}
	}
}

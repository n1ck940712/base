package slack

import (
	"errors"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/api"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type Channel int8

var isEnabled = true

const (
	LootboxTesting Channel = iota
	LootboxHealthCheck
	LOLTower
	LOLCouple
	LOLCoupleMonitor777Dev
	LOLCoupleMonitor777Live
)

const (
	Identifier                     = "slack-message-channel"
	urlHookJus                     = "https://hooks.slack.com/services/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"
	urlHookLootboxSlackTesting     = "https://hooks.slack.com/services/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"
	urlHookLootboxHealthCheck      = "https://hooks.slack.com/services/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"
	urlHookLOLTowerHealthCheck     = "https://hooks.slack.com/services/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"
	urlHookLOLCoupleHealthCheck    = "https://hooks.slack.com/services/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"
	urlHookLOLCoupleMonitor777Dev  = "https://hooks.slack.com/services/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"
	urlHookLOLCoupleMonitor777Live = "https://hooks.slack.com/services/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"
)

func SetEnable(enable bool) {
	isEnabled = enable
}

func SendMessage(title string, msg string, channel Channel) error {
	if !isEnabled {
		return errors.New("slack is disabled")
	}
	if err := generateBaseAPI(channel).AddBody(map[string]any{
		"text": "`" + settings.ENVIRONMENT + "` " + title,
		"attachments": []any{
			map[string]any{
				"color": "#FF0000",
				"text":  "```" + msg + "```",
			},
		},
	}).Post(nil); err != nil {
		return err
	}
	return nil
}

func SendPayload(payload Payload, channel Channel) error {
	if !isEnabled {
		return errors.New("slack is disabled")
	}
	if err := generateBaseAPI(channel).AddBody(payload).Post(nil); err != nil {
		return err
	}
	return nil
}

func generateBaseAPI(channel Channel) api.API {
	return api.NewAPI(generateUrl(channel)).
		SetIdentifier(Identifier + *types.Int(channel).String().Ptr()).
		AddHeaders(map[string]string{
			"Content-type": "application/json",
		})
}

func generateUrl(channel Channel) string {
	switch channel {
	case LootboxTesting:
		return urlHookLootboxSlackTesting
	case LootboxHealthCheck:
		if settings.GetEnvironment().String() == "local" { //override when local use testing instead
			return urlHookLootboxSlackTesting
		}
		return urlHookLootboxHealthCheck
	case LOLTower:
		return urlHookLOLTowerHealthCheck
	case LOLCouple:
		return urlHookLOLCoupleHealthCheck
	case LOLCoupleMonitor777Dev:
		return urlHookLOLCoupleMonitor777Dev
	case LOLCoupleMonitor777Live:
		return urlHookLOLCoupleMonitor777Live
	default:
		return urlHookJus //default
	}
}

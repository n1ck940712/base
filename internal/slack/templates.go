package slack

import (
	"fmt"
	"strings"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
)

func NewLootboxMessage(origin string, environment string, build string, messages ...string) Payload {
	payload := NewPayload().
		AddSection(func(section Section) Section {
			section.AddFieldMarkdown(fmt.Sprintf("*Environment:* %v", strings.ToUpper(environment)))
			section.AddFieldMarkdown(fmt.Sprintf("*Build:* %v", build))
			return section
		})

	if len(messages) > 0 {
		payload.
			AddDivider()

		for i := 0; i < len(messages); i++ {
			payload.AddSection(func(section Section) Section {
				section.SetMarkdown(messages[i])
				return section
			})
		}
	}

	return payload
}

func NewLootboxNotification(origin string, messages ...string) Payload {
	environment := settings.GetEnvironment().String()
	build := settings.GetBuildVersion().String()
	notificationName := environmentIcon(environment) + "*" + strings.ToUpper(origin) + "*"

	return NewLootboxMessage(origin, environment, build, messages...).AsNotification(notificationName, environmentColor(environment))
}

func IdentifierToTitle(identifier string) string {
	switch identifier {
	case constants_loltower.Identifier:
		return "LOL 🗼 "
	case constants_lolcouple.Identifier:
		return "LOL 👫 "
	case constants_fifashootup.Identifier:
		return "Soccer ⚽ "
	case constants_fishprawncrab.Identifier:
		return "Fish Prawn 🦀 "
	default:
		return identifier + " "
	}
}

func environmentIcon(env string) string {
	switch env {
	case "live":
		return "🚀 "
	case "dev":
		return "🚁 "
	default:
		return "🛺 "
	}
}

func environmentColor(env string) string {
	switch env {
	case "live":
		return "#FF0000"
	case "dev":
		return "#0000FF"
	default:
		return "#00FF00"
	}
}

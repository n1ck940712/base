package process_config

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
)

const ConfigType = "config"

type ConfigDatasource interface {
	GetIdentifier() string
	GetUser() *models.User
	GetEvent() *models.Event
	GetGameTable() *models.GameTable
	GetMemberTable() *models.MemberTable
	GetMemberConfigs() *[]models.MemberConfig
	GetBetChips() *[]float64
}

type ConfigProcess interface {
	GetConfig() response.ResponseData
}

type configProcess struct {
	datasource ConfigDatasource
}

func NewConfigProcess(datasource ConfigDatasource) ConfigProcess {
	return &configProcess{datasource: datasource}
}

func (cp *configProcess) GetConfig() response.ResponseData {
	event := cp.datasource.GetEvent()
	memberTable := cp.datasource.GetMemberTable()
	currentEvent := (*response.ConfigEventData)(nil)

	if event != nil {
		currentEvent = &response.ConfigEventData{
			ID:            *event.ID,
			StartDatetime: event.StartDatetime,
		}
	}

	tour := false
	config := response.ConfigData{
		CurrentEvent:    currentEvent,
		BetChips:        cp.datasource.GetBetChips(),
		Enable:          false,
		ResultAnimation: true,
		EffectsSound:    0.5,
		GameSound:       0.5,
		Tour:            &tour,
		IsAnonymous:     false,
	}

	if memberTable != nil && memberTable.IsEnabled {
		config.MinBetAmount = &memberTable.MinBetAmount
		config.MaxBetAmount = &memberTable.MaxBetAmount
		config.MaxPayoutAmount = &memberTable.MaxPayoutAmount
		config.Enable = true
	}

	for _, configItem := range *cp.datasource.GetMemberConfigs() {
		switch configItem.Name {
		case "showCharts":
			config.SetShowCharts(configItem.Value)
		case "result_animation":
			config.SetResultAnimation(configItem.Value)
		case "game_sound":
			config.SetGameSound(configItem.Value)
		case "effects_sound":
			config.SetEffectsSound(configItem.Value)
		case "tour":
			config.SetTour(configItem.Value)
		}
	}

	return &config
}

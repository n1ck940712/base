package process_lolcouple

import (
	"testing"

	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
)

type ltDatasource struct {
}

func (td *ltDatasource) GetUser() *models.User {
	user := models.User{
		ID: 14,
	}

	if err := db.Shared().Where(user).First(&user).Error; err != nil {
		logger.Error("db get user error: ", err.Error())
		return nil
	}
	return &user
}

func TestProcessLOLCouple(t *testing.T) {
	db.UseLocalhost()
	datasource := ltDatasource{} //test datasource
	lolCoupleProcess := NewLOLCoupleProcess(
		process.NewDatasource(constants_lolcouple.Identifier, datasource.GetUser(), constants_lolcouple.GameID, constants_lolcouple.TableID),
	)

	configRequest := utils.PrettyJSON(map[string]any{
		"type": "config",
		"data": nil,
	})
	configResponse := lolCoupleProcess.ProcessRequest(configRequest)
	logger.Info("configResponse: ", utils.PrettyJSON(configResponse))

	oddsRequest := utils.PrettyJSON(map[string]any{
		"type": "odds",
		"data": nil,
	})
	oddsResponse := lolCoupleProcess.ProcessRequest(oddsRequest)
	logger.Info("oddsResponse: ", utils.PrettyJSON(oddsResponse))

	ticketStateRequest := utils.PrettyJSON(map[string]any{
		"type": "current-tickets",
		"data": nil,
	})
	ticketStateResponse := lolCoupleProcess.ProcessRequest(ticketStateRequest)
	logger.Info("ticketStateResponse: ", utils.PrettyJSON(ticketStateResponse))
	betRequest := utils.PrettyJSON(map[string]any{
		"type": "bet",
		"data": map[string]any{
			"event_id": "auto",
			"table_id": "13",
			"tickets": []map[string]any{
				{
					"selection":   "6_under",
					"amount":      10,
					"market_type": 19,
					"selection_data": map[string]any{
						"lower_power":  2000,
						"higher_power": 3050,
					},
					"reference_no": "lolTower_2974377_m_111233",
				},
			},
		}})
	betResponse := lolCoupleProcess.ProcessRequest(betRequest)
	logger.Info("betResponse: ", utils.PrettyJSON(betResponse))
}

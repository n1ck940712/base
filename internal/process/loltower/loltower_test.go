package process_loltower

import (
	"testing"

	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
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

func TestProcessLOLTower(t *testing.T) {
	db.UseLocalhost()
	datasource := ltDatasource{} //test datasource
	lolTowerProcess := NewLOLTowerProcess(
		process.NewDatasource(constants_loltower.Identifier, datasource.GetUser(), constants_loltower.GameID, constants_loltower.TableID),
	)

	configRequest := utils.PrettyJSON(map[string]any{
		"type": "config",
		"data": nil,
	})
	configResponse := lolTowerProcess.ProcessRequest(configRequest)

	logger.Info("configResponse: ", utils.PrettyJSON(configResponse))
	oddsRequest := utils.PrettyJSON(map[string]any{
		"type": "odds",
		"data": nil,
	})
	oddsResponse := lolTowerProcess.ProcessRequest(oddsRequest)

	logger.Info("oddsResponse: ", utils.PrettyJSON(oddsResponse))
	betRequest := utils.PrettyJSON(map[string]any{
		"type": "bet",
		"data": map[string]any{
			"event_id": "auto",
			"table_id": "11",
			"tickets": []map[string]any{
				{
					"selection":   "w",
					"amount":      10,
					"market_type": 18,
					"selection_data": map[string]any{
						"lower_power":  2000,
						"higher_power": 3050,
					},
					"reference_no": "lolTower_2974377_m_111233",
				},
			},
		}})
	lolTowerProcess.SetEventID(807416) //set event id testing
	betResponse := lolTowerProcess.ProcessRequest(betRequest)

	logger.Info("betResponse: ", utils.PrettyJSON(betResponse))
	selection1Request := utils.PrettyJSON(map[string]any{
		"type": "selection",
		"data": map[string]any{
			"selection": "w",
		}})
	lolTowerProcess.SetEventID(807418) //set event id testing
	selection1Response := lolTowerProcess.ProcessRequest(selection1Request)

	logger.Info("selectionResponse: ", utils.PrettyJSON(selection1Response))
	selection2Request := utils.PrettyJSON(map[string]any{
		"type": "selection",
		"data": map[string]any{
			"selection": "l",
		}})
	lolTowerProcess.SetEventID(807419) //set event id testing
	selection2Response := lolTowerProcess.ProcessRequest(selection2Request)

	logger.Info("selection2Response: ", utils.PrettyJSON(selection2Response))
	selection3Request := utils.PrettyJSON(map[string]any{
		"type": "selection",
		"data": map[string]any{
			"selection": 1,
		}})
	lolTowerProcess.SetEventID(807421) //set event id testing
	selection3Response := lolTowerProcess.ProcessRequest(selection3Request)

	logger.Info("selection3Response: ", utils.PrettyJSON(selection3Response))
}

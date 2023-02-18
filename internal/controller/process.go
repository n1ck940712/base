package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/request"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	fifashootup_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/fifashootup"
	fishprawncrab_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/fishprawncrab"
	lolcouple_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/lolcouple"
	loltower_bet "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-bet/loltower"
	process_config "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-config"
	fifashootup_game_data "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-game-data/fifashootup"
	process_member_list "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-member-list"
	fifashootup_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds/fifashootup"
	fishprawncrab_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds/fishprawncrab"
	lolcouple_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds/lolcouple"
	loltower_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds/loltower"
	process_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-state"
	fifashootup_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state/fifashootup"
	fishprawncrab_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state/fishprawncrab"
	lolcouple_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state/lolcouple"
	loltower_ticket_state "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-ticket-state/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/gin-gonic/gin"
)

type ProcessController interface {
	Process(pType string) gin.HandlerFunc
}

type processController struct {
}

func NewProcessController() ProcessController {
	return &processController{}
}

func (pc *processController) Process(pType string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := ValidateTableID(ctx); err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		user, ok := ctx.MustGet("user").(*models.User)

		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user was not passed, please check validate token"})
			return
		}
		var resp response.ResponseData
		requestData := request.NewRequest()

		if err := requestData.ParseJSON(pc.GenerateRequestJSON(pType, ctx)); err != nil {
			resp = response.ErrorBadRequest(pType)
		} else {
			tableID := GetTableID(ctx)
			datasource := process.NewProcessDatasource(process.NewDatasource(GetIdentifier(ctx), user, GetGameID(ctx), GetTableID(ctx)))
			datasource.SetGetEventCallback(func() *models.Event {
				return pc.GetProcessCurrentEvent(tableID)
			})

			switch pType {
			case process.StateType:
				resp = pc.GetState(tableID, datasource, requestData)
			case process.ConfigType:
				resp = pc.GetConfig(tableID, datasource, requestData)
			case process.BetType:
				resp = pc.PlaceBet(tableID, datasource, requestData)
			case process.SelectionType:
				resp = pc.PlaceSelection(tableID, datasource, requestData)
			case process.OddsType:
				resp = pc.GetOdds(tableID, datasource, requestData)
			case process.TicketType, process.CurrentTicketsType:
				resp = pc.GetTicketState(tableID, datasource, requestData)
			case process.MemberListType:
				resp = pc.GetMemberList(tableID, datasource, requestData)
			case process.GameDataType:
				resp = pc.GetGameData(tableID, datasource, requestData)
			default:
				resp = response.ErrorWithMessage(pType+" is not supported", pType)
			}
		}
		if err, ok := resp.(*response.ErrorData); ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		} else {
			ctx.JSON(http.StatusOK, resp)
		}
	}
}

func (pc *processController) GenerateRequestJSON(pType string, ctx *gin.Context) string {
	bytes, _ := io.ReadAll(ctx.Request.Body)
	data := map[string]any{}
	json.Unmarshal(bytes, &data)
	uBytes, _ := json.Marshal(map[string]any{
		"type": pType,
		"data": data,
	})
	return string(uBytes)
}

func (pc *processController) GetState(tableID int64, datasource process.ProcessDatasource, requestData *request.Request) response.ResponseData {
	switch tableID {
	case constants_loltower.TableID,
		constants_lolcouple.TableID,
		constants_fifashootup.TableID,
		constants_fishprawncrab.TableID:
		return process_state.NewStateProcess(datasource).GetState()
	default:
		return response.ErrorWithMessage("state type is not supported", requestData.Type)
	}
}

func (pc *processController) GetOdds(tableID int64, datasource process.ProcessDatasource, requestData *request.Request) response.ResponseData {
	switch tableID {
	case constants_loltower.TableID:
		return loltower_odds.NewOddsProcess(datasource).GetOdds()
	case constants_lolcouple.TableID:
		return lolcouple_odds.NewOddsProcess(datasource).GetOdds()
	case constants_fifashootup.TableID:
		return fifashootup_odds.NewOddsProcess(datasource).GetOdds()
	case constants_fishprawncrab.TableID:
		return fishprawncrab_odds.NewOddsProcess(datasource).GetOdds()
	default:
		return response.ErrorWithMessage("odds type is not supported", requestData.Type)
	}
}

func (pc *processController) GetConfig(tableID int64, datasource process.ProcessDatasource, requestData *request.Request) response.ResponseData {
	switch tableID {
	case constants_loltower.TableID,
		constants_lolcouple.TableID,
		constants_fifashootup.TableID,
		constants_fishprawncrab.TableID:
		return process_config.NewConfigProcess(datasource).GetConfig()
	default:
		return response.ErrorWithMessage("config type is not supported", requestData.Type)
	}
}

func (pc *processController) PlaceBet(tableID int64, datasource process.ProcessDatasource, requestData *request.Request) response.ResponseData {
	lKey := fmt.Sprintf("%v-%v-%v", datasource.GetUser().ID, tableID, requestData.Type)

	if Lock(lKey) {
		return nil
	}
	defer Unlock(lKey)
	switch tableID {
	case constants_loltower.TableID:
		if err := requestData.LOLTowerValidateBet(); err != nil {
			return response.ErrorWithMessage(err.Error(), requestData.Type)
		}
		return loltower_bet.NewBetProcess(datasource).Placebet(requestData.GetBetData())
	case constants_lolcouple.TableID:
		if err := requestData.LOLCoupleValidateBet(); err != nil {
			return response.ErrorWithMessage(err.Error(), requestData.Type)
		}
		return lolcouple_bet.NewBetProcess(datasource).Placebet(requestData.GetBetData())
	case constants_fifashootup.TableID:
		if err := requestData.FIFAShootupValidateBet(); err != nil {
			return response.ErrorWithMessage(err.Error(), requestData.Type)
		}
		return fifashootup_bet.NewBetProcess(datasource).Placebet(requestData.GetBetData())
	case constants_fishprawncrab.TableID:
		if err := requestData.FishPrawnCrabValidateBet(); err != nil {
			return response.ErrorWithMessage(err.Error(), requestData.Type)
		}
		return fishprawncrab_bet.NewBetProcess(datasource).Placebet(requestData.GetBetData())
	default:
		return response.ErrorWithMessage("bet type is not supported", requestData.Type)
	}
}

func (pc *processController) PlaceSelection(tableID int64, datasource process.ProcessDatasource, requestData *request.Request) response.ResponseData {
	lKey := fmt.Sprintf("%v-%v-%v", datasource.GetUser().ID, tableID, requestData.Type)

	if Lock(lKey) {
		return nil
	}
	defer Unlock(lKey)
	switch tableID {
	case constants_loltower.TableID:
		if err := requestData.LOLTowerValidateSelection(); err != nil {
			return response.ErrorWithMessage(err.Error(), requestData.Type)
		}
		return loltower_bet.NewBetProcess(datasource).PlaceSelection(requestData.GetSelectionData())
	case constants_fifashootup.TableID:
		if err := requestData.FIFAShootupValidateSelection(); err != nil {
			return response.ErrorWithMessage(err.Error(), requestData.Type)
		}
		return fifashootup_bet.NewBetProcess(datasource).PlaceSelection(requestData.GetSelectionData())
	default:
		return response.ErrorWithMessage("selection type is not supported", requestData.Type)
	}
}

func (pc *processController) GetTicketState(tableID int64, datasource process.ProcessDatasource, requestData *request.Request) response.ResponseData {
	switch tableID {
	case constants_loltower.TableID:
		return loltower_ticket_state.NewTicketStateProcess(datasource).GetTicketState()
	case constants_lolcouple.TableID:
		return lolcouple_ticket_state.NewTicketStateProcess(datasource).GetTicketState()
	case constants_fifashootup.TableID:
		return fifashootup_ticket_state.NewTicketStateProcess(datasource).GetTicketState()
	case constants_fishprawncrab.TableID:
		return fishprawncrab_ticket_state.NewTicketStateProcess(datasource).GetTicketState()
	default:
		return response.ErrorWithMessage("ticket_state type is not supported", requestData.Type)
	}
}

func (pc *processController) GetMemberList(tableID int64, datasource process.ProcessDatasource, requestData *request.Request) response.ResponseData {
	switch tableID {
	case constants_loltower.TableID:
		return process_member_list.NewMemberListProcess(datasource).GetMemberList()
	default:
		return response.ErrorWithMessage("state type is not supported", requestData.Type)
	}
}

func (pc *processController) GetGameData(tableID int64, datasource process.ProcessDatasource, requestData *request.Request) response.ResponseData {
	switch tableID {
	case constants_fifashootup.TableID:
		return fifashootup_game_data.NewGameDataProcess(datasource).GetGameData()
	default:
		return response.ErrorWithMessage("state type is not supported", requestData.Type)
	}
}

func (pc *processController) GetProcessCurrentEvent(tableID int64) *models.Event {
	event := models.Event{}

	rawQuery := fmt.Sprintf(`
		SELECT
			*
		FROM 
			mini_game_event
		WHERE 
			mini_game_table_id = %v AND
			start_datetime <= NOW()%v
		ORDER BY ctime DESC
		LIMIT 1
		`, tableID, pc.GetProcessMaxEventMS(tableID))

	if err := db.Shared().Preload("Results").Raw(rawQuery).First(&event).Error; err != nil {
		logger.Error("process-route GetEvent error: ", err.Error())
		return nil
	}
	return &event
}

func (pc *processController) GetProcessMaxEventMS(tableID int64) string {
	switch tableID {
	case constants_loltower.TableID:
		return fmt.Sprintf(` AND NOW() < (start_datetime + INTERVAL '%v milliseconds')`, constants_loltower.StartBetMS+constants_loltower.StopBetMS+constants_loltower.ShowResultMS)
	case constants_lolcouple.TableID:
		return fmt.Sprintf(` AND NOW() < (start_datetime + INTERVAL '%v milliseconds')`, constants_lolcouple.StartBetMS+constants_lolcouple.StopBetMS+constants_lolcouple.ShowResultMaxMS)
	case constants_fifashootup.TableID:
		return fmt.Sprintf(` AND NOW() < (start_datetime + INTERVAL '%v milliseconds')`, constants_fifashootup.StartBetMS+constants_fifashootup.StopBetMS+constants_fifashootup.ShowResultMS)
	case constants_fishprawncrab.TableID:
		return fmt.Sprintf(` AND NOW() < (start_datetime + INTERVAL '%v milliseconds')`, constants_fishprawncrab.StartBetMS+constants_fishprawncrab.StopBetMS+constants_fishprawncrab.ShowResultMS)
	default:
		panic("please provide max event ms for tableID " + types.Int(tableID).String())
	}
}
